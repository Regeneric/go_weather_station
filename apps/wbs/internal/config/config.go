package config

import (
	"fmt"
	"os"
	"sx126x"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/uart"
)

type Config struct {
	MQTT    MQTT          `yaml:"mqtt"`
	SPI     SPI           `yaml:"spi"`
	I2C     I2C           `yaml:"i2c"`
	UART    UART          `yaml:"uart"`
	OneWire OneWire       `yaml:"onewire"`
	LoRa    sx126x.Config `yaml:"lora"`
}

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
	Devices map[string]SPIDevice `yaml:"device"`
}

type SPIDevice struct {
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
	Devices map[string]I2CDevice `yaml:"device"`
}

type I2CDevice struct {
	Enable bool   `yaml:"enable" env:"I2C_ENABLE" env-default:"false"`
	Name   string `yaml:"name" env:"I2C_DEVICE" env-default:"0"`
}

// ------------------------------------------------------------------------

// ************************************************************************
// = UART ===
// ------------------------------------------------------------------------
type UART struct {
	Enable  bool                  `yaml:"enable" env:"UART_ENABLE" env-default:"false"`
	Devices map[string]UARTDevice `yaml:"device"`
}

type UARTDevice struct {
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
	Devices map[string]OneWireDevice `yaml:"device"`
}

type OneWireDevice struct {
	Enable bool   `yaml:"enable" env:"ONEWIRE_ENABLE" env-default:"false"`
	Name   string `yaml:"name" env:"ONEWIRE_DEVICE" env-default:"1"`
}

// ------------------------------------------------------------------------

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
