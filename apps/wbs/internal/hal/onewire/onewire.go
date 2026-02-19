package onewire

import (
	"fmt"
	"log/slog"
	"wbs/internal/config"

	ponewire "periph.io/x/conn/v3/onewire"
	ponewirereg "periph.io/x/conn/v3/onewire/onewirereg"
	"periph.io/x/host/v3"
)

func Init(device string) (ponewire.BusCloser, error) {
	log := slog.With("func", "Init()", "params", "(string)", "return", "(onewire.BusCloser, error)", "package", "onewire")
	log.Info("Initializing 1-Wire bus", "device", device)

	// Load drivers for RPi
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[1W] Host init failed: %w", err)
	}

	// /dev/i2c-0 ; /dev/i2c-1 etc.
	bus, err := ponewirereg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[1W] Failed to open 1-Wire bus %s: %w", device, err)
	}

	return bus, nil
}

func Setup(cfg *config.OneWire) (map[string]ponewire.BusCloser, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*config.OneWire)", "return", "(map[string]onewire.BusCloser, func(), error)", "package", "onewire")
	log.Info("1-Wire bus setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("1-Wire bus disabled in the config file")
	}

	conns := make(map[string]ponewire.BusCloser)
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

		bus, err := Init(dev.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to init 1-W %s (%s): %w", key, dev.Name, err)
		}
		closers = append(closers, bus.Close)

		conns[key] = bus
		slog.Debug("1-W bus configured", "key", key, "name", dev.Name)
	}

	return conns, cleanup, nil
}
