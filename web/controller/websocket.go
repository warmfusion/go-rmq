package controller

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"github.com/warmfusion/go-rmq/dto"
	"github.com/warmfusion/go-rmq/qutils"
	"github.com/warmfusion/go-rmq/web/model"
)

type websocketController struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func newWebsocketController(url string) *websocketController {
	log.Println("Starting websocket")

	wsc := new(websocketController)

	wsc.conn, wsc.ch = qutils.GetChannel(url)

	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	go wsc.listenForSources()
	go wsc.listenForMessages()

	return wsc
}

func (wsc *websocketController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)
	go wsc.listenForDiscoveryRequests(socket)
}

func (wsc *websocketController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *websocketController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()

	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)

		}
	}

	wsc.mutex.Unlock()
}

func (wsc *websocketController) listenForSources() {
	q := qutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		qutils.WebappSourceExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	msgs, _ := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	for msg := range msgs {
		log.Printf("Getting details of %s", string(msg.Body))

		sensor, err := model.GetSensorByName(string(msg.Body))
		if err != nil {
			println(err.Error())
		}
		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	}
}

func (wsc *websocketController) listenForMessages() {
	q := qutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		qutils.WebappMetricsExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	msgs, _ := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	for msg := range msgs {
		buf := bytes.NewBuffer(msg.Body)
		dec := gob.NewDecoder(buf)
		sm := dto.SensorMessage{}
		err := dec.Decode(&sm)

		if err != nil {
			println(err.Error())
		}

		wsc.sendMessage(message{
			Type: "reading",
			Data: sm,
		})
	}
}

func (wsc *websocketController) sendMessage(msg message) {
	socketsToRemove := []*websocket.Conn{}

	for _, socket := range wsc.sockets {
		err := socket.WriteJSON(msg)

		if err != nil {
			socketsToRemove = append(socketsToRemove, socket)
		}
	}

	for _, socket := range socketsToRemove {
		wsc.removeSocket(socket)
	}
}

func (wsc *websocketController) listenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			wsc.ch.Publish(
				"", //exchange string,
				qutils.WebappDiscoveryQueue, //key string,
				false,             //mandatory bool,
				false,             //immediate bool,
				amqp.Publishing{}) //msg amqp.Publishing)
		}
	}
}

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
