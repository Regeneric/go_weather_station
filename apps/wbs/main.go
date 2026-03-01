package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sx126x"
	"syscall"
	"wbs/internal/config"
	"wbs/internal/hal/i2c"
	"wbs/internal/hal/onewire"
	"wbs/internal/hal/spi"
	"wbs/internal/hal/uart"
	"wbs/internal/lora"

	"periph.io/x/host/v3"
)

func main() {
	// ************************************************************************
	// = Platform Setup ===
	// ------------------------------------------------------------------------

	// Platform drivers
	if _, err := host.Init(); err != nil {
		panic(err)
	}

	// ctx, cancel := context.WithCancel(context.Background())
	_, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() { <-sigChan; cancel() }() // Wait for Ctrl + C, basically

	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = Logger ===
	// ------------------------------------------------------------------------
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = Config ===
	// ------------------------------------------------------------------------
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Error("Critical error loading configuration", "error", err)
		os.Exit(1)
	}

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		logger.Info("Config file not found, using Defaults & Environment Variables only")
	} else {
		logger.Info("Configuration loaded", "sourceFile", *configPath)
	}

	// cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	// fmt.Printf("Loaded Config:\n%s\n", string(cfgJSON))
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SPI ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	spiConnections, spiClose, err := spi.Setup(&cfg.SPI)
	if err != nil {
		slog.Error("Critical SPI init failure", "error", err)
	} else {
		defer spiClose()
	}

	hkSPI_0, ok := spiConnections["spi0"]
	if !ok {
		slog.Error("Missing SPI device configuration", "name", "spi0")
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = I2C ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	i2cConnections, i2cClose, err := i2c.Setup(&cfg.I2C)
	if err != nil {
		slog.Error("Critical I2C init failure", "error", err)
	} else {
		defer i2cClose()
	}

	hkI2C_0, ok := i2cConnections["i2c0"]
	if !ok {
		slog.Error("Missing I2C bus configuration", "name", "i2c0")
	}

	_ = hkI2C_0 // Temporary
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = 1-Wire ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	owConnections, owClose, err := onewire.Setup(&cfg.OneWire)
	if err != nil {
		slog.Error("Critical 1-W init failure", "error", err)
	} else {
		defer owClose()
	}

	hkOW_0, ok := owConnections["ow0"]
	if !ok {
		slog.Error("Missing 1-W bus configuration", "name", "ow0")
	}

	_ = hkOW_0 // Temporary
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = UART ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	uartConnections, uartClose, err := uart.Setup(&cfg.UART)
	if err != nil {
		slog.Error("Critical UART init failure", "error", err)
	} else {
		defer uartClose()
	}

	hkUART_0, ok := uartConnections["uart0"]
	if !ok {
		slog.Error("Missing UART bus configuration", "name", "uart0")
	}

	_ = hkUART_0 // Temporary
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SX1262 ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	// 15:8 - RFU ; 7:0 - status codes
	var LoraStatusRegister lora.Status

	hkSX1262_0, err := sx126x.New(hkSPI_0, &cfg.SX126X)
	if err != nil || hkSX1262_0 == nil {
		slog.Error("Critical SX126x modem failure", "error", err)
		LoraStatusRegister |= lora.SX126XModemError
	}

	hkLoRa_0, err := lora.New(hkSX1262_0, &cfg.SX126X)
	if err != nil {
		slog.Error("Critical LoRa mode modem failure", "error", err)
		LoraStatusRegister |= lora.LoraModemError
	}

	if err := lora.Setup(hkLoRa_0); err != nil {
		slog.Error("Critical LoRa mode modem setup failure", "error", err)
		LoraStatusRegister |= lora.LoraSetupError
	} else {
		defer hkLoRa_0.Close()
	}
	// ------------------------------------------------------------------------
}
