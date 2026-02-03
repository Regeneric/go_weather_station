package i2c

import (
	"fmt"
	"log/slog"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func Init(device string) (i2c.BusCloser, error) {
	log := slog.With("func", "Init", "params", "string", "package", "bus", "module", "i2c")
	log.Info("Initializing I2C bus", "device", device)

	// Load drivers for RPi
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[I2C] Host init failed: %w", err)
	}

	// /dev/i2c-1 ; /dev/i2c-2 ; /dev/i2c-3 etc.
	bus, err := i2creg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[I2C] Failed to open I2C bus %s: %w", device, err)
	}

	return bus, nil
}
