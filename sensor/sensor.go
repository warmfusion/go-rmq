package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/warmfusion/go-rmq/dto"
	"github.com/warmfusion/go-rmq/qutils"

	"github.com/streadway/amqp"
)

var (
	url      = flag.String("rmq_url", "amqp://guest:guest@192.168.99.100:5672", "RabbitMQ endpoint URL")
	name     = flag.String("name", "sensor", "name of the sensor")
	freq     = flag.Uint("freq", 5, "update frequency in cycles/sec")
	max      = flag.Float64("max", 5., "maximum value for generated readings")
	min      = flag.Float64("min", 1., "minimum value for generated readings")
	stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")
)

// Random numbery ness with seed of UnixNano
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Value = Current sensor value
var value = r.Float64()*(*max-*min) + *min

// Where should the normalised point be (The middle)
var nom = (*max-*min)/2 + *min

func main() {
	// Parse all the input flags
	flag.Parse()

	//Create a new connection and channel against the provided url
	conn, ch := qutils.GetChannel(*url)
	defer conn.Close() // Ensure that the Connection is closed when main() finishes
	defer ch.Close()   // Ensure that the Channel is closed when main() finishes

	// Create/Get the queue to use to send data to
	dataQueue := qutils.GetQueue(*name, ch, false)

	publishQueueName(ch)
	// Create a new QueueListener to await discovery messages
	discoveryQueue := qutils.GetQueue("", ch, true)
	ch.QueueBind(
		discoveryQueue.Name, //name string,
		"",                  //key string,
		qutils.SensorDiscoveryExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	go listenForDiscoveryRequests(discoveryQueue.Name, ch)

	// How long to wait between ticks
	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
	signal := time.Tick(dur)

	// Buffer for encoding the message to be sent
	buf := new(bytes.Buffer)

	// Blocks on signal till the tick count is reached
	for range signal {
		calcValue() // Get a new value (Stored in package variable)

		// Create a new reading measurement
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}

		// Ensure the buffer is clear of any old data
		buf.Reset()
		enc := gob.NewEncoder(buf) // Need to make a new encoder each time
		enc.Encode(reading)        // Encode that sucker into the buffer we've given it

		// Creates a new message struct which we can publish to the queue
		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		// Publish our message to the dataQueue
		ch.Publish(
			"",             //exchange string,
			dataQueue.Name, //key string,
			false,          //mandatory bool,
			false,          //immediate bool,
			msg)            //msg amqp.Publishing)

		log.Printf("Reading sent. Value: %v\n", value)
	}
}

// publishQueueName Announces this sensors queue name
// to the discovery queue
func publishQueueName(ch *amqp.Channel) {
	// Announce the existence of a new Sensor with *name
	// so that the consumers know to start listening to this sensor
	msg := amqp.Publishing{Body: []byte(*name)}
	// Publish to the fanout exchange so any number of consumers get be running
	ch.Publish(
		"amq.fanout", //exchange string,
		"",           //key string,
		false,        //mandatory bool,
		false,        //immediate bool,
		msg)          //msg amqp.Publishing)

}

func listenForDiscoveryRequests(name string, ch *amqp.Channel) {

	msgs, _ := ch.Consume(
		name,  //queue string,
		"",    //consumer string,
		true,  //autoAck bool,
		false, //exclusive bool,
		false, //noLocal bool,
		false, //noWait bool,
		nil)   //args amqp.Table)

	for range msgs {
		log.Print("Received discovery request...")
		publishQueueName(ch)
	}

}

// calcValue will return a value between the max/min values configured
// but tries to keep the values near the midpoint using some weighted
// normals - this makes it look a bit more natural
func calcValue() {
	var maxStep, minStep float64

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += r.Float64()*(maxStep-minStep) + minStep
}
