package main

import (
	"time"

	"github.com/Regeneric/go_weather_station/pkg/model/bme"
	"github.com/Regeneric/go_weather_station/pkg/model/system"
)

const (
	SensorType             = iota // 0
	SensorTemperature             // 1
	SensorPressure                // 2
	SensorRelativeHumidity        // 3
	SensorAbsoluteHumidity        // 4
	SensorDewPoint                // 5
	SensorCO2                     // 6
	SensorTVOC                    // 7

	DataStruct = 99
)

func NormalizeQueue[T any](input <-chan T, output chan<- system.SensorPacket, sourceID string) {
	go func() {
		for data := range input {
			output <- system.SensorPacket{
				SourceID:  sourceID,
				Payload:   data,
				Timestamp: time.Now().UnixMilli(),
			}
		}
	}()
}

type AverageData struct {
	Value  float32
	Weight uint8
}

func WeightedAverage(data []AverageData) float32 {
	var sum, divider float32

	for _, item := range data {
		if item.Weight > 0 {
			sum += item.Value * float32(item.Weight)
			divider += float32(item.Weight)
		}
	}

	if divider == 0 {
		return 0.0
	}

	return sum / divider
}

func GetQueueData(snapshot SystemSnapshot, id string, metric int) interface{} {
	payload, exists := snapshot[id]
	if !exists {
		return nil
	}

	switch data := payload.(type) {
	case bme.Data:
		switch metric {
		case SensorType:
			return SensorType
		case SensorTemperature:
			return data.Temperature
		case SensorRelativeHumidity:
			return data.Humidity
		case SensorAbsoluteHumidity:
			return data.AbsoluteHumidity
		case SensorDewPoint:
			return data.DewPoint
		case SensorPressure:
			return data.Pressure

		case DataStruct:
			return data
		}
		// case dht.Data:
		// case pms.Data:
	}

	return nil
}

func ToFloat(val interface{}) float32 {
	if v, ok := val.(float32); ok {
		return v
	}

	return 0.0
}
