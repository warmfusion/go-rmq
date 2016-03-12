package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"github.com/warmfusion/go-rmq/dto"
	"github.com/warmfusion/go-rmq/qutils"
)

// QueueListener representing our Queue Listener
type QueueListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery
}

// NewQueueListener Get a new QueueListener
func NewQueueListener(url string) *QueueListener {

	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
	}

	ql.conn, ql.ch = qutils.GetChannel(url)

	return &ql
}

// DiscoverSensors sends a fanout message to all connected
// Sensors to report back their queue names so the coordinator
// can start listening to them
func (ql *QueueListener) DiscoverSensors() {
	log.Println("Discovering any existing sensors...")
	//ensure that the Discovery Exchange is setup
	ql.ch.ExchangeDeclare(
		qutils.SensorDiscoveryExchange, //name string,
		"fanout",                       //kind string,
		false,                          //durable bool,
		true,                           //autoDelete bool,
		false,                          //internal bool,
		false,                          //noWait bool,
		nil)                            //args amqp.Table)

	// Push an empty message onto the queue as this'll be enough
	// to indicate to sensors that somethings looking for them
	//
	// TODO: Maybe send a 'response queue' so that we can isolate
	// new coordinator's getting sensor blasts from everyone else?
	ql.ch.Publish(
		qutils.SensorDiscoveryExchange, //exchange string,
		"",                //key string,
		false,             //mandatory bool,
		false,             //immediate bool,
		amqp.Publishing{}) //msg amqp.Publishing)
}

// ListenForNewSource is attached to the QueueListener
// struct (the parans between func and function name)
// so that it can be invoked with ql.ListenForNewSource()
func (ql *QueueListener) ListenForNewSource() {
	q := qutils.GetQueue("", ql.ch)
	ql.ch.QueueBind(
		q.Name,       //name string,
		"",           //key string,
		"amq.fanout", //exchange string,
		false,        //noWait bool,
		nil)          //args amqp.Table)

	// Listen for messages on the fanout Exchange as these represent
	// new sensors coming live that need to be listend to
	msgs, _ := ql.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	// As we're now ready to recieve messages from the sensors
	// lets discover them
	ql.DiscoverSensors()

	for sensor := range msgs {

		if ql.sources[string(sensor.Body)] == nil {
			log.Print(fmt.Sprintf("New Sensor detected: %s", sensor.Body))

			delivery, _ := ql.ch.Consume(
				string(sensor.Body), //queue string,
				"",                  //consumer string,
				true,                //autoAck bool,
				false,               //exclusive bool,
				false,               //noLocal bool,
				false,               //noWait bool,
				nil)                 //args amqp.Table

			ql.sources[string(sensor.Body)] = delivery
			log.Print("Added new sensor to list of sources")
			go ql.AddListener(delivery)

		}
	}

}

// AddListener Watches the amqp.Delivery channel for new messages and handles them
func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
	for msg := range msgs { // Blocks while waiting for new messages
		r := bytes.NewReader(msg.Body) // Read the bosy into a bytes buffer
		d := gob.NewDecoder(r)         // Creates a new decoder off the bytes buffer
		sd := new(dto.SensorMessage)   // Create a new instance of our Struct to decode INTO
		d.Decode(sd)                   // actually decode into a struct of our Sensor Message

		// Output a message of great win
		fmt.Printf("Received message: %v\n", sd)

	}
}
