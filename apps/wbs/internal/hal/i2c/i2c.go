package i2c

import (
	"fmt"
	"log/slog"
	"wbs/internal/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func New(device string) (i2c.BusCloser, error) {
	log := slog.With("func", "New()", "params", "(string)", "return", "(i2c.BusCloser, error)", "package", "i2c")
	log.Info("Initializing I2C bus", "device", device)

	// Load drivers for RPi
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[I2C] Host init failed: %w", err)
	}

	// /dev/i2c-0 ; /dev/i2c-1 etc.
	bus, err := i2creg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[I2C] Failed to open I2C bus %s: %w", device, err)
	}

	return bus, nil
}

func Setup(cfg *config.I2C) (map[string]i2c.BusCloser, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*config.I2C)", "return", "(map[string]i2c.BusCloser, func(), error)", "package", "i2c")
	log.Info("I2C bus setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("I2C bus disabled in the config file")
	}

	conns := make(map[string]i2c.BusCloser)
	var closers []func() error

	cleanup := func() {
		for i, c := range closers {
			slog.Debug("Closing I2C bus connection...", "connection", i)
			_ = c()
		}
	}

	for key, dev := range cfg.Devices {
		if dev.Enable == false {
			continue
		}

		bus, err := New(dev.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to init I2C %s (%s): %w", key, dev.Name, err)
		}
		closers = append(closers, bus.Close)

		conns[key] = bus
		slog.Debug("I2C bus configured", "key", key, "name", dev.Name)
	}

	return conns, cleanup, nil
}
