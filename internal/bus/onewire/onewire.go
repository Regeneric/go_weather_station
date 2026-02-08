package onewire

import (
	"fmt"
	"log/slog"

	"github.com/Regeneric/go_weather_station/internal/config"
	"periph.io/x/conn/v3/onewire"
	"periph.io/x/conn/v3/onewire/onewirereg"
	"periph.io/x/host/v3"
)

func Init(device string) (onewire.BusCloser, error) {
	log := slog.With("func", "Init()", "params", "(string)", "return", "(onewire.BusClose, error)", "package", "bus", "module", "onewire")
	if config.Enable1Wire == false {
		return nil, fmt.Errorf("1-Wire has been disabled in the config file: Enable1Wire == %v", config.Enable1Wire)
	}
	log.Info("Initializing OneWire bus", "bus", device)

	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[OneWire] Host init failed: %w", err)
	}

	bus, err := onewirereg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[OneWire] Failed to open OneWire bus %s: %w", device, err)
	}

	return bus, nil
}
