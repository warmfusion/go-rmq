package main

import (
	"flag"
	"fmt"

	"github.com/warmfusion/go-rmq/coordinator"
)

var (
	url   = flag.String("rmq_url", "amqp://guest:guest@192.168.99.100:5672", "RabbitMQ endpoint URL")
	gHost = flag.String("graphite_host", "localhost", "Host of the Graphite service")
	gPort = flag.Int("graphite_port", 2003, "Port of the Graphite service")
)

var wc *coordinator.WebappConsumer

func main() {
	flag.Parse()

	er := coordinator.NewEventAggregator()
	ql := coordinator.NewQueueListener(er, string(*url))

	graphiteHandler := coordinator.GetGraphiteHandle(*gHost, *gPort)
	coordinator.NewGraphiteConsumer(er, string(*url), graphiteHandler)

	wc = coordinator.NewWebappConsumer(er, string(*url))

	go ql.ListenForNewSource()

	// This will hold the code open till we elect to stop it
	// by opening a blocking read on stdin (no loops, just block)
	var a string
	fmt.Scanln(&a)

}
