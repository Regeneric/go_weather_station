package main

import (
	"log/slog"
	"os"

	"github.com/Regeneric/go_weather_station/internal/bus/i2c"
	"github.com/Regeneric/go_weather_station/internal/bus/onewire"
	"github.com/Regeneric/go_weather_station/internal/bus/uart"
	"github.com/Regeneric/go_weather_station/internal/config"
)

func main() {
	// ************************************************************************
	// = Logger ===
	// ------------------------------------------------------------------------
	opts := &slog.HandlerOptions{
		Level:     config.LogLevel,
		AddSource: true,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = I2C ===
	// ------------------------------------------------------------------------
	hkI2C1, err := i2c.Init(config.HardI2C)
	if err != nil {
		slog.Error("[CRITICAL] Failed to initialized I2C bus", "bus", "i2c-"+config.HardI2C, "err", err)
		os.Exit(1)
	}

	defer func() {
		slog.Debug("Closing I2C bus...", "bus", "i2c-"+config.HardI2C)
		if err := hkI2C1.Close(); err != nil {
			slog.Error("Error closing I2C bus", "bus", "i2c-"+config.HardI2C, "err", err)
		}
	}()

	slog.Debug("I2C bus initialized successfully", "bus", "i2c-"+config.HardI2C)

	hkI2C2, err := i2c.Init(config.SoftI2C)
	if err != nil {
		slog.Error("[CRITICAL] Failed to initialized I2C bus", "bus", "i2c-"+config.SoftI2C, "err", err)
		os.Exit(1)
	}

	defer func() {
		slog.Debug("Closing I2C bus...", "bus", "i2c-"+config.SoftI2C)
		if err := hkI2C2.Close(); err != nil {
			slog.Error("Error closing I2C bus", "bus", "i2c-"+config.SoftI2C, "err", err)
		}
	}()

	slog.Debug("I2C bus initialized successfully", "bus", "i2c-"+config.SoftI2C)
	slog.Info("All I2C buses initialized successfully")
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = UART ===
	// ------------------------------------------------------------------------
	uart.Scan()

	hkUART0, err := uart.Init(config.UARTPort, config.UARTBaudRate)
	if err != nil {
		slog.Error("[CRITICAL] Failed to initialize UART device", "port", config.UARTPort, "err", err)
		// Do not exit on this error, core functions don't need UART
	} else {
		defer func() {
			slog.Debug("Closing UART device...", "port", config.UARTPort)
			if err := hkUART0.Close(); err != nil {
				slog.Error("Error closing UART device", "port", config.UARTPort, "err", err)
			}
		}()

		slog.Debug("UART device initialized successfully", "port", config.UARTPort)
		slog.Info("All UART devices initialized successfully")
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = OneWire ===
	// ------------------------------------------------------------------------
	hkOneWire0, err := onewire.Init(config.Soft1Wire)
	if err != nil {
		slog.Error("[CRITICAL] Failed to initialize OneWire bus", "bus", config.Soft1Wire, "err", err)
		// Do not exit on this error, core functions don't need 1W
	} else {
		defer func() {
			slog.Debug("Closing OneWire bus...", "bus", config.Soft1Wire)
			if err := hkOneWire0.Close(); err != nil {
				slog.Error("Error closing OneWire bus", "bus", config.Soft1Wire, "err", err)
			}
		}()

		slog.Debug("OneWire bus initialized successfully", "bus", config.Soft1Wire)
		slog.Info("All OneWire buses initialized successfully")
	}
	// ------------------------------------------------------------------------
}
