package config

import (
	"log/slog"
	"time"
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
	MQTTBrokerAddress = "192.168.1.50"
	MQTTBrokerPort    = 1883
	MQTTopic          = "sensors/data"
	MQTTKeepAlive     = 60 * time.Second
	MQTTDeviceName    = "pico"
)

// -------------------------------------------------------------------------

// --- Communication -------------------------------------------------------
const (
	HardI2C = "1" // /dev/i2c-1
	SoftI2C = "3" // /dev/i2c-3
)

const (
	UARTPort     = "/dev/serial0"
	UARTBaudRate = 9600
)

const (
	Soft1Wire = "1" // /sys/bus/w1/devices/w1_bus_master1
)

// -------------------------------------------------------------------------
