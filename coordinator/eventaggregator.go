package coordinator

import (
	"log"
	"time"
)

// EventRaiser Generalised interface for raising events
type EventRaiser interface {
	AddListener(name string, f func(interface{}))
}

// EventAggregator contains a map of listeners which have provided callbacks
// to be executed when events are fired
type EventAggregator struct {
	listeners map[string][]func(interface{})
}

// NewEventAggregator Constructor to create new EventAggregator
func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(interface{})),
	}
	return &ea
}

// AddListener appends a callback function to the given event name
func (ea *EventAggregator) AddListener(name string, f func(interface{})) {
	log.Printf("Added listener for %s", name)
	ea.listeners[name] = append(ea.listeners[name], f)
}

// PublishEvent Triggers a new event callback for the given name and eventData
func (ea *EventAggregator) PublishEvent(name string, eventData interface{}) {
	log.Printf("Event published; %s", name)
	if ea.listeners[name] != nil {
		for _, r := range ea.listeners[name] {
			log.Printf("Invoking provided function")
			r(eventData)
		}
	}
}

// EventData - Describes an event
type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}
