package sx1262

import (
	"fmt"
	"log/slog"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
)

func New(connection spi.Conn, cfg *lora.Config) (*LoraModem, error) {
	log := slog.With("func", "New()", "params", "(spi.Conn, *lora.Config)", "return", "(*LoraModem, error)", "package", "comm", "module", "sx1262")
	log.Debug("LoRa modem constructor", "spi", connection, "config", cfg)

	if cfg.Enable == false {
		return nil, fmt.Errorf("LoRa has been disabled in the config file!")
	}

	hw, err := hardwareSetup(connection, cfg)
	if err != nil {
		return nil, fmt.Errorf("LoRa hardware setup failed: %w", err)
	}

	if cfg.RxQueueSize <= 0 {
		cfg.RxQueueSize = 10
		log.Warn("RX queue size cannot be less than 1; resized to 10", "size", cfg.RxQueueSize)
	}

	if cfg.TxQueueSize <= 0 {
		cfg.TxQueueSize = 10
		log.Warn("RX queue size cannot be less than 1; resized to 10", "size", cfg.TxQueueSize)
	}

	modem := LoraModem{
		hw:        hw,
		RxQueue:   make(chan []uint8, cfg.RxQueueSize),
		TxQueue:   make(chan []uint8, cfg.TxQueueSize),
		irqQueue:  make(chan struct{}, 1),
		stopQueue: make(chan struct{}, 1),
	}

	if err := calibrationSequence(&modem); err != nil {
		return nil, err
	}

	return &modem, nil
}

func (d *LoraModem) Close(params ...uint8) error {
	log := slog.With("func", "Close()", "params", "...uint8", "package", "comm", "module", "sx1262")
	log.Info("Shutting down LoRa modem...")

	var sleepType uint8
	if len(params) > 0 {
		sleepType = params[0]
	}

	commands := []uint8{lora.CmdSetSleep, sleepType}
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not shutdown LoRa modem [% X]: %w", commands, err)
	}

	return nil
}

func (d *LoraModem) Run(retries uint8) {
	log := slog.With("func", "Run()", "params", "(uint8)", "return", "(-)", "package", "comm", "module", "sx1262")
	log.Info("Run LoRa modem manager")

	go func() {
		for {
			if d.hw.DIO.WaitForEdge(lora.RxNoTimeout) {
				select {
				case d.irqQueue <- struct{}{}: // 0 bytes struct as a signal, that DIO pin has changed its state
				default: // IRQ queue is full, so we must wait for other routines to handle it anyway
				}
			}
		}
	}()

	if err := d.SetRx(lora.RxContinuous); err != nil {
		log.Error("Could not enable LoRa RX mode", "mode", fmt.Sprintf("% X", lora.RxContinuous), "err", err)
	}

	for {
		select {
		case <-d.irqQueue:
			d.handleIRQ(retries)
		case data := <-d.TxQueue:
			d.dataTransmit(data, d.hw.Config.TxBufferAddress, d.hw.Config.TxTimeout)
		case <-d.stopQueue:
			return
		}
	}
}

func (d *LoraModem) Tx(data []uint8) {
	log := slog.With("func", "Tx()", "params", "([]uint8)", "return", "(-)", "package", "comm", "module", "sx1262")
	log.Debug("Transmit data over LoRa", "data", fmt.Sprintf("% X", data))

	d.TxQueue <- data
}

func hardwareSetup(connection spi.Conn, cfg *lora.Config) (*lora.Device, error) {
	log := slog.With("func", "hardwareSetup()", "params", "(spi.Conn, *lora.Config)", "return", "(*lora.Device, error)", "package", "comm", "module", "sx1262")
	log.Info("Initializing LoRA module")

	sx1262 := lora.Device{
		SPI:    connection,
		Reset:  gpioreg.ByName(cfg.Pins.Reset),
		Busy:   gpioreg.ByName(cfg.Pins.Busy),
		CS:     gpioreg.ByName(cfg.Pins.CS),
		DIO:    gpioreg.ByName(cfg.Pins.DIO),
		TxEn:   gpioreg.ByName(cfg.Pins.TxEn),
		Config: cfg,
	}

	if err := sx1262.CS.Out(gpio.High); err != nil {
		return nil, fmt.Errorf("Failed to set CS pin state to HIGH: %w", err)
	}

	if err := sx1262.Reset.Out(gpio.High); err != nil {
		return nil, fmt.Errorf("Failed to set RESET pin state to HIGH: %w", err)
	}

	if err := sx1262.Busy.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
		return nil, fmt.Errorf("Failed to set BUSY pin edge detection: %w", err)
	}

	if err := sx1262.DIO.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		return nil, fmt.Errorf("Failed to set DIO1 pin pull down and edge detection: %w", err)
	}

	if sx1262.TxEn != nil {
		if err := sx1262.TxEn.Out(gpio.Low); err != nil {
			return nil, fmt.Errorf("Failed to set TxEn pin state to LOW (reciever mode): %w", err)
		}
	}

	return &sx1262, nil
}

func calibrationSequence(modem *LoraModem) error {
	log := slog.With("func", "calibrationSequence()", "params", "(*lora.Modem)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Info("Calibrating LoRA module")

	if err := modem.HardReset(); err != nil {
		return err
	}
	if err := modem.SetStandby(modem.hw.Config.StandbyMode); err != nil {
		return err
	}
	if err := modem.SetPacketType(modem.hw.Config.Modem); err != nil {
		return err
	}
	if err := modem.CalibrateImage(modem.hw.Config.FrequencyRange); err != nil {
		return err
	}
	if err := modem.SetRfFrequency(modem.hw.Config.Frequency); err != nil {
		return err
	}
	if err := modem.SetTxParams(modem.hw.Config.TXPower, modem.hw.Config.RampTime); err != nil {
		return err
	}
	if err := modem.SetModulationParams(modem.hw.Config.SF, modem.hw.Config.Bandwidth, modem.hw.Config.CR, modem.hw.Config.LDRO); err != nil {
		return err
	}
	if err := modem.SetPacketParams(modem.hw.Config.PreambleLen, modem.hw.Config.HeaderType, modem.hw.Config.PayloadLen, modem.hw.Config.CRCType, modem.hw.Config.InvertIQ); err != nil {
		return err
	}
	if err := modem.SetDioIrqParams(modem.hw.Config.IRQMask); err != nil {
		return err
	}
	if err := modem.SetBufferBaseAddress(modem.hw.Config.TxBufferAddress, modem.hw.Config.RxBufferAddress); err != nil {
		return err
	}

	if err := modem.WriteRegister(lora.RegLoraSyncWordMsb, []uint8{uint8(modem.hw.Config.SyncWord >> 8), uint8(modem.hw.Config.SyncWord & 0xFF)}); err != nil {
		return err
	}
	if read, err := modem.ReadRegister(lora.RegLoraSyncWordMsb, 2); err != nil {
		return err
	} else {
		log.Debug("Read register success", "syncWord", fmt.Sprintf("% X", read)) // Simple read test
	}

	return nil
}
