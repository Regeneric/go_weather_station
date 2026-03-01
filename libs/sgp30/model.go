package sgp30

import (
	"periph.io/x/conn/v3/i2c"
)

type Config struct {
	Enable               bool   `yaml:"enable" env:"SGP30_ENABLE" env-default:"false"`
	Name                 string `yaml:"name" env:"SGP30_NAME"`
	HumidityCompensation bool   `yaml:"humidity_compensation" env:"SGP30_HUMIDITY_COMPENSATION" env-default:"false"`
	UseDHT               bool   `yaml:"use_dht" env:"SGP30_USE_DHT" env-default:"false"`
	UseBME               bool   `yaml:"use_bme" env:"SGP30_USE_BME" env-default:"false"`
	Address              uint8  `yaml:"address" env:"SGP30_ADDRESS" env-default:"0x58"`
	Location             string `yaml:"location" env:"SGP30_LOCATION"`
}

type Device struct {
	I2C    i2c.BusCloser
	Config *Config
}
