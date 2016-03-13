package main

import (
	"flag"
	"net/http"

	"github.com/warmfusion/go-rmq/coordinator"
	"github.com/warmfusion/go-rmq/web/controller"
	"github.com/warmfusion/go-rmq/web/model"
)

var (
	url = flag.String("rmq_url", "amqp://guest:guest@192.168.99.100:5672", "RabbitMQ endpoint URL")
)

func main() {
	flag.Parse()
	controller.Initialize(*url)

	er := coordinator.NewEventAggregator()
	ql := coordinator.NewQueueListener(er, string(*url))
	sc := coordinator.NewSourceConsumer(er, string(*url))

	go ql.ListenForNewSource()

	model.Init(sc)

	http.ListenAndServe(":3000", nil)
}
