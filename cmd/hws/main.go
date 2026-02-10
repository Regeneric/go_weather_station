package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Regeneric/go_weather_station/internal/app/mosquitto"
	"github.com/Regeneric/go_weather_station/internal/bus/i2c"
	"github.com/Regeneric/go_weather_station/internal/bus/onewire"
	spis "github.com/Regeneric/go_weather_station/internal/bus/spi"
	"github.com/Regeneric/go_weather_station/internal/bus/uart"
	"github.com/Regeneric/go_weather_station/internal/comm/sx1262"
	"github.com/Regeneric/go_weather_station/internal/config"
	"github.com/Regeneric/go_weather_station/internal/sensors/bme280"
	"github.com/Regeneric/go_weather_station/internal/utils"
	"github.com/Regeneric/go_weather_station/pkg/model/bme"
	"github.com/Regeneric/go_weather_station/pkg/model/lora"
	"github.com/Regeneric/go_weather_station/pkg/model/system"
	"github.com/mcuadros/go-defaults"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/host/v3"
)

func main() {
	// ************************************************************************
	// *** Setup **************************************************************

	// Platform drivers
	if _, err := host.Init(); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() { <-sigChan; cancel() }() // Wait for Ctrl + C, basically

	mqttClient := mosquitto.MQTTInit()

	// ************************************************************************
	// = Logger ===
	// ------------------------------------------------------------------------
	opts := &slog.HandlerOptions{
		Level:     config.LogLevel,
		AddSource: false,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = I2C ===
	// ------------------------------------------------------------------------
	hkI2C1, err := i2c.Init(config.I2COne)
	if err != nil {
		slog.Error("Failed to initialize I2C bus", "bus", "i2c-"+config.I2COne, "err", err)
	} else {
		slog.Debug("I2C bus initialized successfully", "bus", "i2c-"+config.I2COne)

		defer func() {
			slog.Debug("Closing I2C bus...", "bus", "i2c-"+config.I2COne)
			if err := hkI2C1.Close(); err != nil {
				slog.Error("Error closing I2C bus", "bus", "i2c-"+config.I2COne, "err", err)
			}
		}()
	}

	hkI2C2, err := i2c.Init(config.I2CTwo)
	if err != nil {
		slog.Error("Failed to initialize I2C bus", "bus", "i2c-"+config.I2CTwo, "err", err)
	} else {
		slog.Debug("I2C bus initialized successfully", "bus", "i2c-"+config.I2CTwo)

		defer func() {
			slog.Debug("Closing I2C bus...", "bus", "i2c-"+config.I2CTwo)
			if err := hkI2C2.Close(); err != nil {
				slog.Error("Error closing I2C bus", "bus", "i2c-"+config.I2CTwo, "err", err)
			}
		}()
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = UART ===
	// ------------------------------------------------------------------------
	hkUART0, err := uart.Init(config.UARTPort, config.UARTBaudRate)
	if err != nil {
		slog.Error("Failed to initialize UART device", "port", config.UARTPort, "err", err)
	} else {
		slog.Debug("UART device initialized successfully", "port", config.UARTPort)

		defer func() {
			slog.Debug("Closing UART device...", "port", config.UARTPort)
			if err := hkUART0.Close(); err != nil {
				slog.Error("Error closing UART device", "port", config.UARTPort, "err", err)
			}
		}()
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SPI ===
	// ------------------------------------------------------------------------
	hkSPI0, err := spis.Init(config.SPIDevice)
	if err != nil {
		slog.Error("Failed to initialized SPI bus", "bus", "SPI0."+config.SPIDevice, "err", err)
	} else {
		defer func() {
			slog.Debug("Closing SPI bus...", "bus", "SPI0."+config.SPIDevice)
			if err := hkSPI0.Close(); err != nil {
				slog.Error("Error closing SPI bus", "bus", "SPI0."+config.SPIDevice, "err", err)
			}
		}()

		slog.Debug("SPI bus initialized successfully", "bus", "SPI0."+config.SPIDevice)
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = OneWire ===
	// ------------------------------------------------------------------------
	hkOneWire0, err := onewire.Init(config.Soft1Wire)
	if err != nil {
		slog.Error("Failed to initialize OneWire bus", "bus", config.Soft1Wire, "err", err)
	} else {
		defer func() {
			slog.Debug("Closing OneWire bus...", "bus", config.Soft1Wire)
			if err := hkOneWire0.Close(); err != nil {
				slog.Error("Error closing OneWire bus", "bus", config.Soft1Wire, "err", err)
			}
		}()

		slog.Debug("OneWire bus initialized successfully", "bus", config.Soft1Wire)
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

	// ************************************************************************
	// = SX1262 ===
	// ------------------------------------------------------------------------
	loraConnection, err := hkSPI0.Connect(config.SPISpeed, spi.Mode0, 8)
	if err != nil {
		slog.Warn("Failed to establish SPI connection to LoRa module. Only local sensors will be available", "err", err)
	}

	hkSX1262_0_GPIO := lora.Pins{
		Reset: config.SX1262ResetPin,
		Busy:  config.SX1262BusyPin,
		DIO:   config.SX1262DIO1Pin,
		CS:    config.SX1262CSPin,
		TxEn:  config.SX1262TxEnPin,
		// RxEn:  config.SX1262RxEnPin,
	}

	hkSX1262_0_Config := lora.Config{
		Enable:         config.SX1262Enable,
		Frequency:      config.SX1262Frequency,
		FrequencyRange: []uint8{lora.CalImg430, lora.CalImg440}, // HACK - for now
		Bandwidth:      config.SX1262Bandwidth,
		SF:             config.SX1262SpreadingFactor,
		CR:             config.SX1262CodingRate,
		LDRO:           config.SX1262LowDataRateOptimize,
		DC_DC:          config.SX1262DC_DC,
		PreambleLen:    config.SX1262PreambleLength,
		PayloadLen:     config.SX1262PayloadLength,
		CRCType:        config.SX1262CRCType,
		InvertIQ:       config.SX1262InvertIQ,
		SyncWord:       config.SX1262SyncWord,
		TXPower:        config.SX1262TransmitPower,
		IRQMask:        lora.IrqTxDone | lora.IrqRxDone | lora.IrqTimeout | lora.IrqCrcErr | lora.IrqHeaderErr, // Hack - for now
		Pins:           &hkSX1262_0_GPIO,
	}
	defaults.SetDefaults(&hkSX1262_0_Config)

	hkSX1262_0, err := sx1262.New(loraConnection, &hkSX1262_0_Config)
	if err != nil {
		slog.Warn("Failed to initialize LoRa module. Only local sensors will be available", "err", err)
	}
	defer hkSX1262_0.Close()
	// ------------------------------------------------------------------------

	// - System Settings ------
	systemState := make(utils.SystemSnapshot)
	inputQueue := make(chan system.SensorPacket, 50)
	outputQueue := make(chan utils.SystemSnapshot, 5)

	processTicker := time.NewTicker(2 * time.Second) // 2 Hz system wide polling rate
	defer processTicker.Stop()

	go mosquitto.MQTTPublish(outputQueue, mqttClient) // - Process Data ----
	// ------------------------

	// - SX1262 ----
	go hkSX1262_0.Run(config.SX1262MaxRetryOnError)
	// -------------

	// - BME280 ----
	utils.NormalizeQueue(hkBME280_0_Data, inputQueue, "BME280_0")
	utils.NormalizeQueue(hkBME280_1_Data, inputQueue, "BME280_1")

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
			snapshot := make(utils.SystemSnapshot)
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
