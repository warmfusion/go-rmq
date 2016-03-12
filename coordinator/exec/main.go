package main

import (
	"flag"
	"fmt"

	"github.com/warmfusion/go-rmq/coordinator"
)

var (
	url = flag.String("rmq_url", "amqp://guest:guest@192.168.99.100:5672", "RabbitMQ endpoint URL")
)

func main() {
	flag.Parse()

	ql := coordinator.NewQueueListener(string(*url))
	go ql.ListenForNewSource()

	// This will hold the code open till we elect to stop it
	// by opening a blocking read on stdin (no loops, just block)
	var a string
	fmt.Scanln(&a)

}
