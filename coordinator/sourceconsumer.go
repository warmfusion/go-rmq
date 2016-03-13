package coordinator

import (
	"fmt"
	"log"

	"github.com/marpaia/graphite-golang"
	"github.com/streadway/amqp"
	"github.com/warmfusion/go-rmq/dto"
	"github.com/warmfusion/go-rmq/qutils"
)

// SourceConsumer Simply used to track the sources we know about
type SourceConsumer struct {
	er              EventRaiser        // Event raiser
	graphiteHandler *graphite.Graphite // Graphite handler
	conn            *amqp.Connection
	ch              *amqp.Channel
	sources         map[string]*dto.Sensor
}

// NewSourceConsumer Create a new consumer
func NewSourceConsumer(er EventRaiser, url string) *SourceConsumer {
	sc := SourceConsumer{
		er:      er,
		sources: make(map[string]*dto.Sensor),
	}

	sc.conn, sc.ch = qutils.GetChannel(url)

	sc.er.AddListener("SensorRegistered", func(event interface{}) {
		sensor := event.(*dto.Sensor)
		sc.AddSensor(sensor)
	})

	return &sc
}

// AddSensor Add any new event sources
func (sc *SourceConsumer) AddSensor(sensor *dto.Sensor) {
	log.Printf("Adding sensor to SourceConsumer [%s]", sensor.Name)
	// Check if we're already subscribed to this event source
	for k := range sc.sources {
		if k == sensor.Name {
			log.Println("Already seen this source; not adding again")
			return // Already watching this one
		}
	}
	sc.sources[sensor.Name] = sensor
}

// GetSensor Find a sensor with the given name
func (sc *SourceConsumer) GetSensor(name string) (*dto.Sensor, error) {
	// Check if we're already subscribed to this event source
	if sc.sources[name] == nil {
		return nil, fmt.Errorf("Sensor [%s] not found", name)
	}

	return sc.sources[name], nil
}
