package process

// import (
// 	"log/slog"

// 	"github.com/Regeneric/go_weather_station/internal/sensors/bme280"
// 	"github.com/Regeneric/go_weather_station/pkg/model/bme"
// )

// func ProcessSensorData(data <-chan SystemSnapshot) {
// 	// Temperature and humidity
// 	// BME280  - weight 10
// 	// DS18B20 - weight 10
// 	// AHT21   - weight 10
// 	// AHT20/DHT20 - weight 9
// 	// DHT22  - weight 7
// 	// BMP180 - weight 3
// 	// DHT11  - weight 2

// 	// Pressure
// 	// BME280 - weight 8
// 	// BMP180 - weight 2

// 	// CO2
// 	// ENS160 - weight 9
// 	// SGP30  - weight 4

// 	// Infinite loop, basically
// 	for snapshot := range data {
// 		slog.Debug("Processing system snapshot", "items", len(snapshot))

// 		// - BME280 ----
// 		BME280_SensorsData := make([]*bme.Data, 0)

// 		if val := GetQueueData(snapshot, "BME280_0", DataStruct); val != nil {
// 			if sensorData, ok := val.(bme.Data); ok {
// 				BME280_SensorsData = append(BME280_SensorsData, &sensorData)
// 			}
// 		}
// 		if val := GetQueueData(snapshot, "BME280_1", DataStruct); val != nil {
// 			if sensorData, ok := val.(bme.Data); ok {
// 				BME280_SensorsData = append(BME280_SensorsData, &sensorData)
// 			}
// 		}

// 		BME280_Avg := bme280.AverageData(BME280_SensorsData)
// 		// -------------

// 		temperatureAvg := []AverageData{
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_0", SensorTemperature)), Weight: 10},
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_1", SensorTemperature)), Weight: 10},
// 			// {Sensor: "DS18B20_0", Value: GetValue(snapshot, "DS18B20_0", SensorTemperature), Weight: 10},
// 		}

// 		dewPointAvg := []AverageData{
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_0", SensorDewPoint)), Weight: 10},
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_1", SensorDewPoint)), Weight: 10},
// 		}

// 		relativeHumidityAvg := []AverageData{
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_0", SensorRelativeHumidity)), Weight: 10},
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_1", SensorRelativeHumidity)), Weight: 10},
// 		}

// 		absoluteHumidityAvg := []AverageData{
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_0", SensorAbsoluteHumidity)), Weight: 10},
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_1", SensorAbsoluteHumidity)), Weight: 10},
// 		}

// 		pressureAvg := []AverageData{
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_0", SensorPressure)), Weight: 10},
// 			{Value: ToFloat(GetQueueData(snapshot, "BME280_1", SensorPressure)), Weight: 10},
// 		}
// 	}
// }
