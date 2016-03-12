package dto

import (
	"encoding/gob"
	"time"
)

// SensorMessage representing a single packet
// of sensor information
type SensorMessage struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

type Sensor struct {
	Name         string  `json:"name"`
	SerialNo     string  `json:"serialNo"`
	UnitType     string  `json:"unitType"`
	MinSafeValue float64 `json:"minSafeValue"`
	MaxSafeValue float64 `json:"maxSafeValue"`
}

func init() {
	// Tell Go that the SensorMessage can be serialised into
	// a string using the 'gob' encoding (Base64 Binary encoding)
	gob.Register(SensorMessage{})
	gob.Register(Sensor{})
}
