package spis

import (
	"fmt"
	"log/slog"

	"github.com/Regeneric/go_weather_station/internal/config"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

func Init(device string) (spi.PortCloser, error) {
	log := slog.With("func", "Init()", "params", "(string)", "return", "(spi.PortCloser, error)", "package", "bus", "module", "spis")
	if config.SPIEnable == false {
		return nil, fmt.Errorf("SPI has been disabled in the config file: SPIEnable == %v", config.SPIEnable)
	}
	log.Info("Initializing SPI bus", "device", device)

	// Load drivers for RPi
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[SPI] Host init failed: %w", err)
	}

	// SPI0.0 ; SPI0.0 etc.
	bus, err := spireg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[SPI] Failed to open SPI bus %s: %w", device, err)
	}

	return bus, nil
}
