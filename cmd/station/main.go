package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Regeneric/go_weather_station/internal/bus/i2c"
	"github.com/Regeneric/go_weather_station/internal/bus/onewire"
	"github.com/Regeneric/go_weather_station/internal/config"
	"github.com/Regeneric/go_weather_station/internal/sensors/bme280"
	"github.com/Regeneric/go_weather_station/pkg/model/bme"
	"github.com/Regeneric/go_weather_station/pkg/model/system"
	"periph.io/x/host/v3"
)

type SystemSnapshot map[string]interface{}

func main() {
	// ************************************************************************
	// *** Setup **************************************************************

	// Load platform drivers
	if _, err := host.Init(); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() { <-sigChan; cancel() }() // Wait for Ctrl + C, basically

	mqttClient := MQTTInit()

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
	// hkUART0, err := uart.Init(config.UARTPort, config.UARTBaudRate)
	// if err != nil {
	// 	slog.Error("[CRITICAL] Failed to initialize UART device", "port", config.UARTPort, "err", err)
	// 	// Do not exit on this error, core functions don't need UART
	// } else {
	// 	defer func() {
	// 		slog.Debug("Closing UART device...", "port", config.UARTPort)
	// 		if err := hkUART0.Close(); err != nil {
	// 			slog.Error("Error closing UART device", "port", config.UARTPort, "err", err)
	// 		}
	// 	}()

	// 	slog.Debug("UART device initialized successfully", "port", config.UARTPort)
	// 	slog.Info("All UART devices initialized successfully")
	// }
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

	// ************************************************************************
	// = BME280 ===
	// ------------------------------------------------------------------------
	hkBME280_0 := bme.Config{
		Address:                    config.BME280Address,
		Status:                     bme.Init,
		HumiditySampling:           bme.HumidOversamplingX2,
		IIRCoefficient:             (bme.StandbyTime0_5ms | bme.FilterCoeff4),
		TemperatureAndPressureMode: (bme.OversamplingX2<<5 | bme.OversamplingX2<<2 | bme.ModeForced),
	}
	hkBME280_0_Data := make(chan bme.Data, 10)
	hkBME280_0_Params := bme.Params{
		I2C:    hkI2C1,
		Config: &hkBME280_0,
	}

	hkBME280_1 := bme.Config{
		Address:                    config.BME280Address,
		Status:                     bme.Init,
		HumiditySampling:           bme.HumidOversamplingX2,
		IIRCoefficient:             (bme.StandbyTime0_5ms | bme.FilterCoeff4),
		TemperatureAndPressureMode: (bme.OversamplingX2<<5 | bme.OversamplingX2<<2 | bme.ModeForced),
	}
	hkBME280_1_Data := make(chan bme.Data, 10)
	hkBME280_1_Params := bme.Params{
		I2C:    hkI2C2,
		Config: &hkBME280_1,
	}
	// ------------------------------------------------------------------------

	// - System Settings ------
	systemState := make(SystemSnapshot)
	inputQueue := make(chan system.SensorPacket, 50)
	outputQueue := make(chan SystemSnapshot, 5)

	processTicker := time.NewTicker(2 * time.Second) // 2 Hz system wide polling rate
	defer processTicker.Stop()

	go MQTTPublish(outputQueue, mqttClient) // - Process Data ----
	// ------------------------

	// - BME280 ----
	NormalizeQueue(hkBME280_0_Data, inputQueue, "BME280_0")
	NormalizeQueue(hkBME280_1_Data, inputQueue, "BME280_1")

	go bme280.Run(ctx, &hkBME280_0_Params, hkBME280_0_Data)
	go bme280.Run(ctx, &hkBME280_1_Params, hkBME280_1_Data)
	// -------------

	// - Data Pipeline ---
	slog.Info("System ready, waiting for data")
	for {
		select {
		// - Collect Data ----
		case packet := <-inputQueue:
			slog.Debug("New sensor data", "source", packet.SourceID, "payload", packet.Payload, "timestamp", packet.Timestamp)
			systemState[packet.SourceID] = packet

		case <-processTicker.C:
			snapshot := make(SystemSnapshot)
			for i, v := range systemState {
				snapshot[i] = v
			}

			select {
			case outputQueue <- snapshot:
			default:
				slog.Warn("Process queue is full")
			}
		// -------------------

		case <-ctx.Done():
			slog.Info("Stopping main loop...")
			return
		}
	}
	// -------------------
}
