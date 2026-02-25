package sx126x

import (
	"fmt"
	"log/slog"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
)

func New(conn spi.Conn, cfg *Config) (*Device, error) {
	log := slog.With("func", "New()", "params", "(spi.Conn, *Config)", "return", "(*Device, error)", "lib", "sx1262")
	log.Info("Initializing SX126x module")

	if cfg.Enable == false {
		return nil, fmt.Errorf("SX126x modem disabled in the config")
	}

	if conn == nil {
		return nil, fmt.Errorf("SPI bus connection state improper")
	}

	loadPin := func(name string) (gpio.PinIO, error) {
		p := gpioreg.ByName(name)
		if p == nil {
			return nil, fmt.Errorf("Pin not found: %s", name)
		}
		return p, nil
	}

	var err error
	pins := &pinsDirection{}

	if pins.reset, err = loadPin(cfg.Pins.Reset); err != nil {
		return nil, err
	}
	if pins.busy, err = loadPin(cfg.Pins.Busy); err != nil {
		return nil, err
	}
	if pins.dio, err = loadPin(cfg.Pins.DIO); err != nil {
		return nil, err
	}
	if pins.txEn, err = loadPin(cfg.Pins.TxEn); err != nil {
		return nil, err
	}
	if cfg.Pins.CS != "" {
		if pins.cs, err = loadPin(cfg.Pins.CS); err != nil {
			return nil, err
		}
	}

	if err := pins.reset.Out(gpio.High); err != nil {
		return nil, fmt.Errorf("Failed to set RESET pin state to HIGH: %w", err)
	}
	if err := pins.busy.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		return nil, fmt.Errorf("Failed to set BUSY pin edge detection: %w", err)
	}
	if err := pins.dio.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		return nil, fmt.Errorf("Failed to set DIO1 pin pull down and edge detection: %w", err)
	}
	if err := pins.txEn.Out(gpio.Low); err != nil {
		return nil, fmt.Errorf("Failed to set TxEn pin state to LOW: %w", err)
	}
	if pins.rxEn != nil {
		if err := pins.rxEn.Out(gpio.Low); err != nil {
			return nil, fmt.Errorf("Failed to set RxEn pin state to LOW: %w", err)
		}
	}
	if pins.cs != nil {
		if err := pins.cs.Out(gpio.High); err != nil {
			return nil, fmt.Errorf("Failed to set CS pin state to HIGH: %w", err)
		}
	}

	if cfg.RxQueueSize <= 0 {
		cfg.RxQueueSize = 10
		log.Warn("RX queue size cannot be less than 1; resized to 10", "size", cfg.RxQueueSize)
	}

	if cfg.TxQueueSize <= 0 {
		cfg.TxQueueSize = 10
		log.Warn("TX queue size cannot be less than 1; resized to 10", "size", cfg.TxQueueSize)
	}

	queue := Queue{
		Rx: make(chan []uint8, cfg.RxQueueSize),
		Tx: make(chan []uint8, cfg.TxQueueSize),
	}

	return &Device{
		SPI:    conn,
		Config: cfg,
		Queue:  queue,
		gpio:   pins,
	}, nil
}

func (d *Device) Close(sleepMode SleepConfig) error {
	log := slog.With("func", "Device.Close()", "params", "(-)", "return", "(error)", "lib", "sx1262")
	log.Info("Closing SX126x module", "mode", sleepMode)

	var err error = nil
	if err = d.SetSleep(sleepMode); err != nil {
		log.Error("Could not set sleep mode", "mode", sleepMode, "error", err)
	}

	if d.gpio.txEn != nil {
		if err = d.gpio.txEn.Out(gpio.Low); err != nil {
			log.Error("Could not set TxEn pin to LOW", "error", err)
		}
	}

	close(d.Queue.Rx)
	close(d.Queue.Tx)

	return err
}
