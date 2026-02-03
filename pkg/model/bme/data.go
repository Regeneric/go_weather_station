package bme

type CalibrationData struct {
	DigT1 uint16
	DigT2 int16
	DigT3 int16

	DigP1 uint16
	DigP2 int16
	DigP3 int16
	DigP4 int16
	DigP5 int16
	DigP6 int16
	DigP7 int16
	DigP8 int16
	DigP9 int16

	DigH1 uint8
	DigH2 int16
	DigH3 uint8
	DigH4 int16
	DigH5 int16
	DigH6 int8
}

type Config struct {
	RawData                    []uint8
	Address                    uint8
	Status                     uint16
	HumiditySampling           uint8
	IIRCoefficient             uint8
	TemperatureAndPressureMode uint8
	Params                     CalibrationData
	Queue                      chan Data
}

type Data struct {
	Sensor           string  `json:"sensor"`
	Pressure         float32 `json:"pressure"`
	Temperature      float32 `json:"temperature"`
	DewPoint         float32 `json:"dewPoint"`
	Humidity         float32 `json:"humidity"`
	AbsoluteHumidity float32 `json:"absoluteHumidity"`
	Timestamp        int64   `json:"timestamp"`
}
