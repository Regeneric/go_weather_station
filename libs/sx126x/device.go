package sx126x

import (
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/spi"
)

type Config struct {
	Enable          bool     `yaml:"enable" env:"SX126X_ENABLE" env-default:"false"`
	Modem           string   `yaml:"modem" env:"SX126X_MODEM" env-default:"lora"`
	Type            string   `yaml:"type" env:"SX126X_TYPE" env-default:"1262"`
	Bandwidth       uint64   `yaml:"bandwidth" env:"SX126X_BANDWIDTH" env-default:"125000"`
	SpreadingFactor uint8    `yaml:"spreading_factor" env:"SX126X_SF" env-default:"7"`
	CodingRate      uint8    `yaml:"coding_rate" env:"SX126X_CR" env-default:"5"`
	LDRO            bool     `yaml:"ldro" env:"SX126X_LDRO" env-default:"false"`
	DC_DC           bool     `yaml:"dc_dc" env:"SX126X_DC_DC" env-default:"false"`
	HeaderImplicit  bool     `yaml:"header_implicit" env:"SX126X_HEADER_IMPLICIT" env-default:"false"`
	Frequency       uint64   `yaml:"frequency" env:"SX126X_FREQUENCY" env-default:"433000000"`
	PreambleLength  uint16   `yaml:"preamble_length" env:"SX126X_PREAMBLE_LEN" env-default:"12"`
	PayloadLength   uint8    `yaml:"payload_length" env:"SX126X_PAYLOAD_LEN" env-default:"32"`
	CRC             bool     `yaml:"crc" env:"SX126X_CRC" env-default:"true"`
	InvertedIQ      bool     `yaml:"inverted_iq" env:"SX126X_IQ_INVERTED" env-default:"false"`
	SyncWord        uint16   `yaml:"sync_word" env:"SX126X_SYNC_WORD" env-default:"5156"` // Aka 0x1424
	TransmitPower   int8     `yaml:"tx_power" env:"SX126X_TX_POWER" env-default:"0"`
	StandbyMode     string   `yaml:"standby_mode" env:"SX126X_STANDBY_MODE" env-default:"rc"`
	SleepMode       string   `yaml:"sleep_mode" env:"SX126X_SLEEP_MODE" env-default:"cold_start"`
	FrequencyRange  []uint16 `yaml:"frequency_range" env:"SX126X_FREQ_RANGE" env-default:"430,440" env-separator:","`
	RampTime        uint16   `yaml:"ramp_time" env:"SX126X_RAMP_TIME" env-default:"800"`
	DIO2AsRfSwitch  bool     `yaml:"dio2_as_rf_switch" env:"SX126X_DIO2_AS_RF_SWITCH" env-default:"true"`
	RxQueueSize     uint8    `yaml:"rx_queue_size" env:"SX126X_RX_QUEUE_SIZE" env-default:"10"`
	TxQueueSize     uint8    `yaml:"tx_queue_size" env:"SX126X_TX_QUEUE_SIZE" env-default:"10"`
	RxBufferAddress uint8    `yaml:"rx_buffer_address" env:"SX126X_RX_BUFFER_ADDRESS" env-default:"128"`
	TxBufferAddress uint8    `yaml:"tx_buffer_address" env:"SX126X_TX_BUFFER_ADDRESS" env-default:"0"`
	TransmitTimeout uint32   `yaml:"tx_timeout" env:"SX126X_TX_TIMEOUT" env-default:"0"`
	Pins            *Pins    `yaml:"pins"`
	CAD             *CAD     `yaml:"cad"`
}

type Pins struct {
	Reset string `yaml:"reset" env:"SX126X_GPIO_RESET" env-default:"GPIO18"`
	Busy  string `yaml:"busy" env:"SX126X_GPIO_BUSY" env-default:"GPIO20"`
	DIO   string `yaml:"dio" env:"SX126X_GPIO_DIO" env-default:"GPIO16"`
	TxEn  string `yaml:"tx_enable" env:"SX126X_GPIO_TX_EN" env-default:"GPIO6"`
	RxEn  string `yaml:"rx_enable" env:"SX126X_GPIO_RX_EN"`
	CS    string `yaml:"cs" env:"SX126X_GPIO_CS"`
}

type CAD struct {
	SymbolNumber     uint8  `yaml:"symbol_number" env:"SX126X_CAD_SYMBOL_NUMBER" env-default:"2"`
	DetectionPeak    uint8  `yaml:"detection_peak" env:"SX126X_CAD_DETECTION_PEAK" env-default:"20"`
	DetectionMinimum uint8  `yaml:"detection_minimum" env:"SX126X_CAD_DETECTION_MINIMUM" env-default:"10"`
	ExitMode         uint8  `yaml:"exit_mode" env:"SX126X_CAD_EXIT_MODE" env-default:"0"`
	Timeout          uint32 `yaml:"timeout" env:"SX126X_CAD_TIMEOUT" env-default:"0"`
}

type pinsDirection struct {
	reset gpio.PinOut
	busy  gpio.PinIn
	dio   gpio.PinIn
	txEn  gpio.PinOut
	rxEn  gpio.PinOut
	cs    gpio.PinOut
}

type ModemStatus struct {
	Command  uint8
	ChipMode uint8
}

type BufferStatus struct {
	RXPayloadLength uint8
	RXStartPointer  uint8
}

type PacketStats struct {
	TotalReceived uint16
	CrcErrors     uint16
	HeaderErrors  uint16
}

type PacketStatus struct {
	SignalStrength         int8
	SnRRatio               float32
	DenoisedSignalStrength int8
	Stats                  PacketStats
}

type Status struct {
	Modem  ModemStatus
	Buffer BufferStatus
	Packet PacketStatus
	Error  DeviceError
}

type Queue struct {
	Rx chan []uint8
	Tx chan []uint8
}

type Device struct {
	SPI    spi.Conn
	Config *Config
	Status Status
	Queue  Queue
	gpio   *pinsDirection
}
