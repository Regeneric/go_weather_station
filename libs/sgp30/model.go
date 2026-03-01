package sgp30

type Config struct {
	Enable  bool           `yaml:"enable" env:"SGP30_ENABLE" env-default:"false"`
	Devices map[string]cfg `yaml:"device"`
}

type cfg struct {
	Enable               bool   `yaml:"enable" env:"SGP30_ENABLE" env-default:"false"`
	Name                 string `yaml:"name" env:"SGP30_NAME"`
	BusName              string `yaml:"bus" env:"SGP30_BUS" env-default:"0"`
	HumidityCompensation bool   `yaml:"humidity_compensation" env:"SGP30_HUMIDITY_COMPENSATION" env-default:"false"`
	UseDHT               bool   `yaml:"use_dht" env:"SGP30_USE_DHT" env-default:"false"`
	UseBME               bool   `yaml:"use_bme" env:"SGP30_USE_BME" env-default:"false"`
	Address              uint8  `yaml:"address" env:"SGP30_ADDRESS" env-default:"0x58"`
	Location             string `yaml:"location" env:"SGP30_LOCATION"`
}

type Bus interface {
	Tx(addr uint16, w, r []uint8) error
	String() string
}

type Device struct {
	I2C    Bus
	Config *cfg
}

type Sensor struct {
	hw *Device
}
