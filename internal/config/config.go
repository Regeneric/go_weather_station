package config

import (
	"log/slog"
	"time"

	"periph.io/x/conn/v3/physic"
)

// --- Log Levels ----------------------------------------------------------
const (
	LogTrace slog.Level = slog.LevelDebug - 4 // -8
	LogDebug slog.Level = slog.LevelDebug     // -4
	LogInfo  slog.Level = slog.LevelInfo      // 0
	LogWarn  slog.Level = slog.LevelWarn      // 4
	LogError slog.Level = slog.LevelError     // 8
)

const LogLevel = LogDebug // Default log level
// -------------------------------------------------------------------------

// --- MQTT ----------------------------------------------------------------
const (
	MQTTEnable            = false
	MQTTBrokerAddress     = "127.0.0.1"
	MQTTBrokerPort        = "1883"
	MQTTopic              = "hws/sensors"
	MQTTKeepAlive         = 60 * time.Second
	MQTTReconnectInterval = 10 * time.Second
	MQTTAutoReconnect     = true
	MQTTDeviceName        = "rpi_zero_0"
)

// -------------------------------------------------------------------------

// --- Communication -------------------------------------------------------
const (
	I2CEnable = false
	HardI2C   = "1" // /dev/i2c-1
	SoftI2C   = "3" // /dev/i2c-3
)

const (
	UARTEnable   = false
	UARTPort     = "/dev/ttyAMA0"
	UARTBaudRate = 9600
)

const (
	Enable1Wire = false
	Soft1Wire   = "1" // /sys/bus/w1/devices/w1_bus_master1
)

const (
	SPIEnable = true
	SPIDevice = "0"
	SPISpeed  = 10 * physic.MegaHertz // 10 MHz
)

// -------------------------------------------------------------------------

// --- BME280 --------------------------------------------------------------
const (
	BME280Enable  = false
	BME280UseI2C  = true
	BME280Address = 0x76
)

// -------------------------------------------------------------------------

// --- SX1262 --------------------------------------------------------------
const (
	SX1262Enable              = true
	SX1262Frequency           = 433 * physic.MegaHertz // 433 MHz
	SX1262Bandwidth           = 500 * physic.KiloHertz // 125 kHz
	SX1262SpreadingFactor     = 7                      // SF
	SX1262CodingRate          = 5                      // CR (5 -> 4/5, 6 -> 4/6, 7 -> 4/7, 8 -> 4/8)
	SX1262LowDataRateOptimize = false                  // LDRO
	SX1262DC_DC               = true                   // DC-DC
	SX1262PreambleLength      = 12
	SX1262PayloadLength       = 32
	SX1262CRCType             = true
	SX1262InvertIQ            = false
	SX1262SyncWord            = 0x1424 // 0x3444 - public network ; 0x1424 - private network
	SX1262TransmitPower       = 0
	SX1262TxEnPin             = "GPIO6"
	SX1262DIO1Pin             = "GPIO16"
	SX1262ResetPin            = "GPIO18"
	SX1262BusyPin             = "GPIO20"
	SX1262CSPin               = "GPIO21"
)

// -------------------------------------------------------------------------
