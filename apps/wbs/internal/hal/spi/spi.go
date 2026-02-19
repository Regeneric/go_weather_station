package spi

import (
	"fmt"
	"log/slog"
	"wbs/internal/config"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

func Init(device string) (spi.PortCloser, error) {
	log := slog.With("func", "Init()", "params", "(string)", "return", "(spi.PortCloser, error)", "package", "spi")
	log.Info("Initializing SPI bus", "device", device)

	// Load drivers for RPi
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[SPI] Host init failed: %w", err)
	}

	// SPI0.0 ; SPI0.1 etc.
	bus, err := spireg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[SPI] Failed to open SPI bus %s: %w", device, err)
	}

	return bus, nil
}

func Setup(cfg *config.SPI) (map[string]spi.Conn, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*config.SPI)", "return", "(map[string]spi.Conn, func(), error)", "package", "spi")
	log.Info("SPI bus setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("SPI bus disabled in the config")
	}

	conns := make(map[string]spi.Conn)
	var closers []func() error

	cleanup := func() {
		for i, c := range closers {
			slog.Debug("Closing SPI bus connection...", "connection", i)
			_ = c()
		}
	}

	for key, dev := range cfg.Devices {
		// Port init
		if dev.Enable == false {
			continue
		}

		port, err := Init(dev.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to init SPI %s (%s): %w", key, dev.Name, err)
		}
		closers = append(closers, port.Close)

		// Connection params
		conn, err := port.Connect(physic.Frequency(dev.Speed), dev.Mode, dev.BitsPerWord)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to configure SPI %s (%s): %w", key, dev.Name, err)
		}

		conns[key] = conn
		slog.Debug("SPI device configured", "key", key, "name", dev.Name, "speed", dev.Speed, "mode", dev.Mode, "bitsPerWord", dev.BitsPerWord)
	}

	return conns, cleanup, nil
}
