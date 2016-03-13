package model

import (
	"github.com/warmfusion/go-rmq/coordinator"
	"github.com/warmfusion/go-rmq/dto"
)

var sourceConsumer *coordinator.SourceConsumer

// Init Create new model
func Init(sc *coordinator.SourceConsumer) {
	sourceConsumer = sc
}

// GetSensorByName go get the sensor data
func GetSensorByName(name string) (*dto.Sensor, error) {
	return sourceConsumer.GetSensor(name)
}
