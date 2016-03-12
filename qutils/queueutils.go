package qutils

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// SensorDiscoveryExchange queue name for auto-discovery of sensors
// The Coordinators request all sensors to reply to a discovery call
// so that the coordinators know they exist
const SensorDiscoveryExchange = "SensorDiscovery"
const WebappSourceExchange = "WebappSourceExchange"
const WebappMetricsExchange = "WebappMetricsExchange" // To get actual live data
const WebappDiscoveryQueue = "WebappDiscoveryQueue"   // To find the sources that are avaliable

// GetChannel Returns a channel for a given AMQP endpoint
func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to establish connection to message broker")
	ch, err := conn.Channel()
	failOnError(err, "Failed to get channel for connection")

	return conn, ch
}

// GetQueue Returns a Queue with the given name
func GetQueue(name string, ch *amqp.Channel, autoDelete bool) *amqp.Queue {
	q, err := ch.QueueDeclare(
		name,       //name string,
		false,      //durable bool,
		autoDelete, //autoDelete bool,
		false,      //exclusive bool,
		false,      //noWait bool,
		nil)        //args amqp.Table)

	failOnError(err, "Failed to declare queue")

	return &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
