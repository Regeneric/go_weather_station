package bme

import (
	"periph.io/x/conn/v3/i2c"
)

type Status int

const (
	Init            = 1
	ReadReady       = 2
	ReadSuccess     = 3
	ReadInProgress  = 4
	ReadBadChecksum = 5
)

type Control int

const (
	RegCalib00  = 0x88
	RegCalib26  = 0xE1
	RegCtrlHum  = 0xF2
	RegCtrlMeas = 0xF4
	RegConfig   = 0xF5

	HumidOversamplingSkipped = 0x00
	HumidOversamplingX1      = 0x01
	HumidOversamplingX2      = 0x02
	HumidOversamplingX4      = 0x03
	HumidOversamplingX8      = 0x04
	HumidOversamplingX16     = 0x05

	OversamplingSkipped = 0x00
	OversamplingX1      = 0x01
	OversamplingX2      = 0x02
	OversamplingX4      = 0x03
	OversamplingX8      = 0x04
	OversamplingX16     = 0x05

	FilterCoeffOff = (0x00 << 2)
	FilterCoeff2   = (0x01 << 2)
	FilterCoeff4   = (0x02 << 2)
	FilterCoeff8   = (0x03 << 2)
	FilterCoeff16  = (0x04 << 2)

	StandbyTime0_5ms  = 0x00
	StandbyTime62_5ms = 0x01
	StandbyTime125ms  = 0x02
	StandbyTime250ms  = 0x03
	StandbyTime500ms  = 0x04
	StandbyTime1000ms = 0x05
	StandbyTime10ms   = 0x06
	StandbyTime20ms   = 0x07
)

type Mode int

const (
	ModeSleep  = 0x00
	ModeForced = 0x01
	ModeNormal = 0x03
)

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
	RawData                    [8]uint8
	Address                    uint16
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

type Params struct {
	I2C    i2c.BusCloser
	Config *Config
}
