package config

import (
	"fmt"
	"os"
	"time"

	"github.com/Regeneric/iot-drivers/libs/sx126x"

	"github.com/Regeneric/iot-drivers/libs/sgp30"

	"github.com/ilyakaznacheev/cleanenv"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/uart"
)

type Config struct {
	Logging Logging       `yaml:"logging"`
	MQTT    MQTT          `yaml:"mqtt"`
	SPI     SPI           `yaml:"spi"`
	I2C     I2C           `yaml:"i2c"`
	UART    UART          `yaml:"uart"`
	OneWire OneWire       `yaml:"onewire"`
	SX126X  sx126x.Config `yaml:"sx126x"`
	BME280  BME280        `yaml:"bme280"`
	DHT     DHT           `yaml:"dht"`
	DS18B20 DS18B20       `yaml:"ds18b20"`
	PMS5003 PMS5003       `yaml:"pms5003"`
	SGP30   sgp30.Group   `yaml:"sgp30"`
}

// ************************************************************************
// = Logging ===
// ------------------------------------------------------------------------
type Logging struct {
	LogLevel string `yaml:"log_level"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = MQTT ===
// ------------------------------------------------------------------------
type MQTT struct {
	Enable            bool          `yaml:"enable" env:"MQTT_ENABLE" env-default:"false"`
	BrokerAddress     string        `yaml:"broker_address" env:"MQTT_BROKER_ADDRESS" env-default:"localhost"`
	BrokerPort        uint16        `yaml:"broker_port" env:"MQTT_BROKER_PORT" env-default:"1883"`
	Topic             []string      `yaml:"topic" env:"MQTT_TOPICS" env-default:"example/topic" env-separator:","`
	KeepAlive         time.Duration `yaml:"keep_alive" env:"MQTT_KEEP_ALIVE" env-default:"60s"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval" env:"MQTT_RECONNECT_INTERVAL" env-default:"10s"`
	AutoReconnect     bool          `yaml:"auto_reconnect" env:"MQTT_AUTO_RECONNECT" env-default:"true"`
	DeviceName        string        `yaml:"device_name" env:"MQTT_DEVICE_NAME" env-default:"example"`
	Username          string        `yaml:"username" env:"MQTT_USERNAME"` // Empty by default
	Password          string        `yaml:"password" env:"MQTT_PASSWORD"` // Empty by default
}

// ------------------------------------------------------------------------

// ************************************************************************
// = SPI ===
// ------------------------------------------------------------------------
type SPI struct {
	Enable  bool                 `yaml:"enable" env:"SPI_ENABLE" env-default:"false"`
	Devices map[string]spiDevice `yaml:"device"`
}

type spiDevice struct {
	Enable      bool     `yaml:"enable" env:"SPI_ENABLE" env-default:"false"`
	Name        string   `yaml:"name" env:"SPI_DEVICE" env-default:"0"`
	Speed       uint64   `yaml:"speed" env:"SPI_SPEED" env-default:"10000000"`
	Mode        spi.Mode `yaml:"mode" env:"SPI_MODE" env-default:"0"`
	BitsPerWord int      `yaml:"bits_per_word" env:"SPI_BITS_PER_WORD" env-default:"8"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = I2C ===
// ------------------------------------------------------------------------
type I2C struct {
	Enable  bool                 `yaml:"enable" env:"I2C_ENABLE" env-default:"false"`
	Devices map[string]i2cDevice `yaml:"device"`
}

type i2cDevice struct {
	Enable bool   `yaml:"enable" env:"I2C_ENABLE" env-default:"false"`
	Name   string `yaml:"name" env:"I2C_DEVICE" env-default:"0"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = UART ===
// ------------------------------------------------------------------------
type UART struct {
	Enable  bool                  `yaml:"enable" env:"UART_ENABLE" env-default:"false"`
	Devices map[string]uartDevice `yaml:"device"`
}

type uartDevice struct {
	Enable     bool      `yaml:"enable" env:"UART_ENABLE" env-default:"false"`
	Name       string    `yaml:"name" env:"UART_DEVICE" env-default:"0" env-separator:","`
	Speed      uint64    `yaml:"speed" env:"UART_SPEED" env-default:"9600"`
	DataLength int       `yaml:"data_length" env:"UART_DATA_LENGTH" env-default:"8"`
	Parity     string    `yaml:"parity_bit" env:"UART_PARITY_BIT" env-default:"N"`
	StopBit    uart.Stop `yaml:"stop_bit" env:"UART_STOP_BIT" env-default:"1"`
	DataFlow   uint8     `yaml:"data_flow" env:"UART_DATA_FLOW" env-default:"0"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = 1-Wire ===
// ------------------------------------------------------------------------
type OneWire struct {
	Enable  bool                     `yaml:"enable" env:"ONEWIRE_ENABLE" env-default:"false"`
	Devices map[string]onewireDevice `yaml:"device"`
}

type onewireDevice struct {
	Enable bool   `yaml:"enable" env:"ONEWIRE_ENABLE" env-default:"false"`
	Name   string `yaml:"name" env:"ONEWIRE_DEVICE" env-default:"1"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = BME280 === TO BE REMOVED FROM HERE
// ------------------------------------------------------------------------
type BME280 struct {
	Enable  bool                    `yaml:"enable" env:"BME280_ENABLE" env-default:"false"`
	Devices map[string]bme280Device `yaml:"device"`
}

type bme280Device struct {
	Enable   bool   `yaml:"enable" env:"BME280_ENABLE" env-default:"false"`
	Name     string `yaml:"name" env:"BME280_DEVICE"`
	UseI2C   bool   `yaml:"use_i2c" env:"BME280_USE_I2C" env-default:"true"`
	Address  uint8  `yaml:"address" env:"BME280_ADDRESS" env-default:"0x76"`
	Location string `yaml:"location" env:"BME280_LOCATION"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = DHT === TO BE REMOVED FROM HERE
// ------------------------------------------------------------------------
type DHT struct {
	Enable  bool                 `yaml:"enable" env:"DHT_ENABLE" env-default:"false"`
	Devices map[string]dhtDevice `yaml:"device"`
}

type dhtDevice struct {
	Enable   bool   `yaml:"enable" env:"DHT_ENABLE" env-default:"false"`
	Name     string `yaml:"name" env:"DHT_NAME"`
	Type     uint8  `yaml:"type" env:"DHT_TYPE" env-default:"20"`
	Address  uint8  `yaml:"address" env:"DHT_ADDRESS" env-default:"56"` // Aka 0x38
	Location string `yaml:"location" env:"DHT_LOCATION"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = DS18B20 === TO BE REMOVED FROM HERE
// ------------------------------------------------------------------------
type DS18B20 struct {
	Enable  bool                     `yaml:"enable" env:"DS18B20_ENABLE" env-default:"false"`
	Devices map[string]ds18b20Device `yaml:"device"`
}

type ds18b20Device struct {
	Enable   bool   `yaml:"enable" env:"DS18B20_ENABLE" env-default:"false"`
	Name     string `yaml:"name" env:"DS18B20_NAME"`
	Location string `yaml:"location" env:"DS18B20_LOCATION"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = PMS5003 === TO BE REMOVED FROM HERE
// ------------------------------------------------------------------------
type PMS5003 struct {
	Enable  bool                     `yaml:"enable" env:"PMS5003_ENABLE" env-default:"false"`
	Devices map[string]pms5003Device `yaml:"device"`
}

type pms5003Device struct {
	Enable               bool   `yaml:"enable" env:"PMS5003_ENABLE" env-default:"false"`
	Name                 string `yaml:"name" env:"PMS5003_NAME"`
	HumidityCompensation bool   `yaml:"humidity_compensation" env:"PMS5003_HUMIDITY_COMPENSATION" env-default:"false"`
	UseDHT               bool   `yaml:"use_dht" env:"PMS5003_USE_DHT" env-default:"false"`
	UseBME               bool   `yaml:"use_bme" env:"PMS5003_USE_BME" env-default:"false"`
	NormalizeData        bool   `yaml:"normalize_data" env:"PMS5003_NORMALIZE_DATA" env-default:"false"`
	Location             string `yaml:"location" env:"PMS5003_LOCATION"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, fmt.Errorf("Config file not found and failed to read ENV: %w", err)
		}

		return cfg, nil
	}

	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file '%s': %w", path, err)
	}

	return cfg, nil
}
