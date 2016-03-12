package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/marpaia/graphite-golang"
	"github.com/streadway/amqp"
	"github.com/warmfusion/go-rmq/qutils"
)

// Idea
// * read event
// * convert to sample
// * Format as Graphite Output
// * ? send to graphite?
// * profit

// GraphiteConsumer Defines a type to read sensor events off the EventRaiser
// queues, converts them into Metrics Events so that they can be handled
// by dedicated handlers
type GraphiteConsumer struct {
	er              EventRaiser        // Event raiser
	graphiteHandler *graphite.Graphite // Graphite handler
	conn            *amqp.Connection
	ch              *amqp.Channel
	sources         []string
}

// NewGraphiteConsumer Create a new GraphiteConsumer
func NewGraphiteConsumer(er EventRaiser, url string, graphiteHandler *graphite.Graphite) *GraphiteConsumer {
	gc := GraphiteConsumer{
		er:              er,
		graphiteHandler: graphiteHandler,
	}

	gc.conn, gc.ch = qutils.GetChannel(url)

	gc.er.AddListener("SensorRegistered", func(eventData interface{}) {
		log.Print("Sensor Registration called" + eventData.(string))
		gc.SubscribeToEventData(eventData.(string))
	})

	return &gc
}

// GetGraphiteHandle Create and return a graphite handle for the given endpoint
func GetGraphiteHandle(host string, port int) *graphite.Graphite {
	// try to connect a graphite server
	g, err := graphite.NewGraphite(host, port)
	// if you couldn't connect to graphite, use a nop
	if err != nil {
		g = graphite.NewGraphiteNop(host, port)
	}

	log.Printf("Loaded Graphite connection: %#v", g)
	g.SimpleSend("stats.graphite_loaded", "1")

	return g
}

// SubscribeToEventData Add any new event sources to the event raiser listeners
func (gc *GraphiteConsumer) SubscribeToEventData(name string) {

	// Check if we're already subscribed to this event source
	for _, v := range gc.sources {
		if v == name {
			return // Already listening - ignore it
		}
	}

	// Hngg... I _think_ I understand whats going on here..
	// We're registering a new callback listener for the given name
	// but rather than simply giving a function to invoke, we're giving a
	// closure which when invokes returns another function (which is what we
	// actually want to be invokved)
	// This lets me share some state (buffer) between calls
	gc.er.AddListener("datum_"+name, func() func(interface{}) {

		buf := new(bytes.Buffer)

		return func(eventData interface{}) {
			ed := eventData.(EventData)

			log.Printf("Metric:: [%s %f %d]", ed.Name, ed.Value, ed.Timestamp.Unix())

			valueAsString := fmt.Sprintf("%f", ed.Value)
			// Create a new reading measurement
			metric := graphite.Metric{
				Name:      ed.Name,
				Value:     valueAsString,
				Timestamp: ed.Timestamp.Unix(),
			}

			gc.graphiteHandler.SendMetric(metric)

			// Ensure the buffer is clear of any old data
			buf.Reset()
			enc := gob.NewEncoder(buf) // Need to make a new encoder each time
			enc.Encode(metric)         // Encode that sucker into the buffer we've given it

			//metric is now ready to be re-emitted to the broker
			gc.ch.Publish(
				"metrics", //exchange string,
				"",        //key string,
				false,     //mandatory bool,
				false,     //immediate bool,
				amqp.Publishing{Body: buf.Bytes()}) //msg amqp.Publishing)
		}
	}())
}
