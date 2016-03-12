package coordinator

import (
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

	// Add the listener callback that'll be fired when a new event is found
	gc.er.AddListener("datum_"+name, gc.handleEvent)
}

// handleEvent Accepts the incoming event and handles it accordingly
func (gc *GraphiteConsumer) handleEvent(eventData interface{}) {

	ed := eventData.(EventData)

	log.Printf("Metric:: [%s %f %d]", ed.Name, ed.Value, ed.Timestamp.Unix())

	// 'Cast' the float to a string
	valueAsString := fmt.Sprintf("%f", ed.Value)
	// Create a new reading measurement
	metric := graphite.Metric{
		Name:      ed.Name,
		Value:     valueAsString,
		Timestamp: ed.Timestamp.Unix(),
	}

	// Send the data to Graphite
	gc.graphiteHandler.SendMetric(metric)
}
