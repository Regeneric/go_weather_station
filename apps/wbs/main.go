package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wbs/internal/config"
	"wbs/internal/hal/i2c"
	"wbs/internal/hal/spi"
	"wbs/internal/lora"
	sgp_manager "wbs/internal/sensors/sgp30"

	"github.com/Regeneric/iot-drivers/libs/sgp30"
	"github.com/Regeneric/iot-drivers/libs/sx126x"

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

	ctx, cancel := context.WithCancel(context.Background())
	// _, cancel := context.WithCancel(context.Background())

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
		logger.Error("[ MAIN ] Critical error loading configuration", "error", err)
		os.Exit(1)
	}

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		logger.Info("[ MAIN ] Config file not found, using Defaults & Environment Variables only")
	} else {
		logger.Info("[ MAIN ] Configuration loaded", "sourceFile", *configPath)
	}

	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("[ MAIN ] Loaded Config:\n%s\n", string(cfgJSON))
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SPI ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	spiConnections, spiClose, err := spi.Setup(&cfg.SPI)
	if err != nil {
		slog.Error("[ MAIN ] Critical SPI init failure", "error", err)
	} else {
		defer spiClose()
	}

	hkSPI_0, ok := spiConnections["spi0"]
	if !ok {
		slog.Error("[ MAIN ] Missing SPI device configuration", "name", "spi0")
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = I2C ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	i2cConnections, i2cClose, err := i2c.Setup(&cfg.I2C)
	if err != nil {
		slog.Error("[ MAIN ] Critical I2C init failure", "error", err)
	} else {
		defer i2cClose()
	}

	_, ok = i2cConnections["i2c1"]
	if !ok {
		slog.Error("[ MAIN ] Missing I2C device configuration", "name", "i2c")
	}
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SX1262 ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	sxlog := lora.SlogAdapter{Log: logger}
	pinreg := lora.PinReg{}

	hkSX1262_0, err := sx126x.New(hkSPI_0, &cfg.SX126X, sx126x.WithLogger(sxlog), sx126x.WithPinReg(pinreg))
	if err != nil || hkSX1262_0 == nil {
		slog.Error("[ MAIN ] Critical SX126x modem failure", "error", err)
	}

	hkLoRa_0, err := lora.New(hkSX1262_0, &cfg.SX126X)
	if err != nil {
		slog.Error("[ MAIN ] Critical LoRa mode modem failure", "error", err)
	}

	if err := lora.Setup(hkLoRa_0); err != nil {
		slog.Error("[ MAIN ] Critical LoRa mode modem setup failure", "error", err)
	} else {
		defer hkLoRa_0.Close()
	}

	go hkLoRa_0.Run(ctx)
	// ------------------------------------------------------------------------

	// ************************************************************************
	// = SGP30 ===  TODO: automatic reconnects after failure
	// ------------------------------------------------------------------------
	sgplog := sgp_manager.SlogAdapter{Log: logger}

	sgp30Buses := make(map[string]sgp30.Bus)
	for key, val := range i2cConnections {
		sgp30Buses[key] = val
	}

	sgp30Sensors, sgp30Close, err := sgp30.Setup(sgp30Buses, &cfg.SGP30, sgplog)
	if err != nil {
		slog.Error("[ MAIN ] Critical SGP30 sensor failure", "error", err)
	} else {
		defer sgp30Close()
	}

	hkSGP30_0, ok := sgp30Sensors["sgp30_0"]
	if !ok {
		slog.Error("[ MAIN ] Missing SGP30 sensor", "name", "sgp30_0")
	}

	if hkSGP30_0 != nil {
		slog.Warn("[ MAIN ] Waiting 20 seconds for SGP30 sensor to initialize...")
		time.Sleep(20 * time.Second)
	}

	hkSGP_PRIMARY := sgp_manager.SGP{HW: hkSGP30_0}
	go hkSGP_PRIMARY.Run(ctx)
	// ------------------------------------------------------------------------

	time.Sleep(1 * time.Second)
	hkLoRa_0.Tx([]uint8("Hello, world!"))
	time.Sleep(2 * time.Second)

	state := "idle"
	for {
		switch state {
		case "idle":
			slog.Debug("[ MAIN ] State Machine", "state", state)

			slog.Info("[ MAIN ] SGP_PRIMARY", "ECO2", hkSGP_PRIMARY.ECO2)
			slog.Info("[ MAIN ] SGP_PRIMARY", "TVOC", hkSGP_PRIMARY.TVOC)

			data, err := hkLoRa_0.Rx(2 * time.Second)
			if err != nil {
				slog.Debug(err.Error()) // It's not really an error
			}

			if len(data) > 0 {
				state = "data_ready"
			}
		case "data_ready":
			slog.Debug("[ MAIN ] State Machine", "state", state)

			slog.Error("[ MAIN ] ABCD")
			state = "idle"
		}
	}
}
