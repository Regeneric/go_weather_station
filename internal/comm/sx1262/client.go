package sx1262

import (
	"fmt"
	"log/slog"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
)

type LoraModem struct {
	hw *lora.Device
}

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

	modem := &LoraModem{hw: hw}
	if err := modem.HardReset(); err != nil {
		return nil, err
	}
	if err := modem.SetStandby(cfg.StandbyMode); err != nil {
		return nil, err
	}
	if err := modem.SetPacketType(cfg.Modem); err != nil {
		return nil, err
	}
	if err := modem.CalibrateImage(cfg.FrequencyRange); err != nil {
		return nil, err
	}
	if err := modem.SetRfFrequency(cfg.Frequency); err != nil {
		return nil, err
	}
	if err := modem.SetTxParams(cfg.TXPower, cfg.RampTime); err != nil {
		return nil, err
	}
	if err := modem.SetModulationParams(cfg.SF, cfg.Bandwidth, cfg.CR, cfg.LDRO); err != nil {
		return nil, err
	}
	if err := modem.SetPacketParams(cfg.PreambleLen, cfg.HeaderType, cfg.PayloadLen, cfg.CRCType, cfg.InvertIQ); err != nil {
		return nil, err
	}
	if err := modem.SetDioIrqParams(cfg.IRQMask); err != nil {
		return nil, err
	}

	if err := modem.WriteRegister(lora.RegLoraSyncWordMsb, []uint8{uint8(cfg.SyncWord >> 8), uint8(cfg.SyncWord & 0xFF)}); err != nil {
		return nil, err
	}
	if read, err := modem.ReadRegister(lora.RegLoraSyncWordMsb, 2); err != nil {
		return nil, err
	} else {
		log.Info("Read register success", "syncWord", fmt.Sprintf("% X", read))
	}

	return modem, nil
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
