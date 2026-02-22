package pms5003

import "wbs/internal/config"

type Sensor struct{}

func New(cfg *config.PMS5003) (Sensor, error) {
	return Sensor{}, nil
}
