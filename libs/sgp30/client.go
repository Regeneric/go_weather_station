package sgp30

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
)

func New(i2c Bus, cfg *cfg) (*Sensor, error) {
	log := slog.With("func", "New()", "params", "(Bus, *cfg)", "return", "(*Sensor, error)", "package", "sgp30")
	log.Info("SGP30 single sensor constructor", "name", cfg.Name)

	if cfg == nil {
		return nil, fmt.Errorf("SGP30 sensor state improper; cfg is nil")
	}
	if i2c == nil || reflect.ValueOf(i2c).IsNil() {
		return nil, fmt.Errorf("SGP30 sensor state improper; i2c is nil")
	}

	return &Sensor{
		hw: &Device{
			I2C:    i2c,
			Config: cfg,
		},
	}, nil
}

// TODO full initilization path with optional settings
func Setup(buses map[string]Bus, cfg *Config) (map[string]*Sensor, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*cfg, map[string]Bus)", "return", "(map[string]*Sensor, func(), error)", "package", "sgp30")
	log.Info("SGP30 sensors setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("SGP30 sensors disabled in the config file")
	}

	sensors := make(map[string]*Sensor)
	var closers []func() error

	cleanup := func() {
		for i, c := range closers {
			log.Debug("Closing SGP30 sensor and saving baseline values...", "sensor", i)
			_ = c()
		}
	}

	for key, dev := range cfg.Devices {
		if dev.Enable == false {
			log.Debug("Sensor disabled in the config file", "name", dev.Name, "bus", dev.BusName, "address", fmt.Sprintf("[%#x]", dev.Address))
			continue
		}

		bus, ok := buses[dev.BusName]
		if !ok {
			cleanup()
			return nil, func() {}, fmt.Errorf("I2C bus '%s' not found for SGP30 sensor '%s' with address '[%#x]'", dev.BusName, dev.Name, dev.Address)
		}

		sensor, err := New(bus, &dev)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}

		if err := sensor.hw.IaqInit(); err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("SGP30 sensors '%s' on bus '%s' and address '[%#x]' unresponsive: %w", dev.Name, dev.BusName, dev.Address, err)
		}

		closers = append(closers, func() error {
			baseline := make([]uint8, 6)
			if err := sensor.hw.GetIaqBaseline(baseline); err != nil {
				return err
			}

			filename := fmt.Sprintf("sgp30_baseline_%s.bin", key)
			if err := os.WriteFile(filename, baseline, 0644); err != nil {
				return err
			}

			return nil
		})

		sensors[key] = sensor
		log.Debug("SGP30 sensor configured and initilized", "name", dev.Name, "bus", dev.BusName, "address", fmt.Sprintf("[%#x]", dev.Address))
	}

	return sensors, cleanup, nil
}
