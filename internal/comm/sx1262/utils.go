package sx1262

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
	"periph.io/x/conn/v3/gpio"
)

func (d *LoraModem) HardReset(timeout ...<-chan time.Time) error {
	log := slog.With("func", "LoraModem.HardReset()", "params", "(..<-chan time.Time)", "package", "comm", "module", "sx1262")
	log.Debug("Hardware reset LoRA module")

	if err := d.hw.CS.Out(gpio.High); err != nil {
		return fmt.Errorf("Failed to set CS pin state to HIGH: %w", err)
	}
	if err := d.hw.Reset.Out(gpio.Low); err != nil {
		return fmt.Errorf("Failed to set RESET pin state to LOW: %w", err)
	}
	time.Sleep(1 * time.Millisecond)
	if err := d.hw.Reset.Out(gpio.High); err != nil {
		return fmt.Errorf("Failed to set RESET pin state to HIGH: %w", err)
	}

	wait := time.After(5 * time.Second)
	if len(timeout) > 0 {
		wait = timeout[0]
	}

	if err := d.BusyCheck(wait); err != nil {
		return fmt.Errorf("Failed to reset LoRa module: %w", err)
	}

	log.Info("LoRa module reset success")
	return nil
}

func (d *LoraModem) BusyCheck(timeout <-chan time.Time, sleep ...time.Duration) error {
	log := slog.With("func", "LoraModem.BusyCheck()", "params", "(<-chan time.Time, ...time.Duration)", "package", "comm", "module", "sx1262")
	log.Debug("Check LoRa module busy status")

	if d.hw.Busy.Read() == gpio.Low {
		log.Debug("LoRa module ready")
		return nil
	}

	interval := 10 * time.Millisecond
	if len(sleep) > 0 {
		interval = sleep[0]
		log.Debug("Sleep interval changed", "interval", interval)
	}

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout")
		default:
			if d.hw.Busy.Read() == gpio.Low {
				log.Debug("LoRa module ready")
				return nil
			}
			time.Sleep(interval) // Avoid busy wait in loop
		}
	}
}

func (d *LoraModem) Write(w []uint8, r []uint8, timeout ...<-chan time.Time) error {
	log := slog.With("func", "LoraModem.Tx()", "params", "([]uint8, []uint8, ...<-chan time.Time)", "package", "comm", "module", "sx1262")
	log.Debug("Send data to LoRa modem")

	wait := time.After(1 * time.Second)
	if len(timeout) > 0 {
		wait = timeout[0]
	}

	if err := d.BusyCheck(wait); err != nil {
		return fmt.Errorf("LoRa module busy: %w", err)
	}

	if err := d.hw.CS.Out(gpio.Low); err != nil {
		return fmt.Errorf("Failed to set CS pin state to LOW: %w", err)
	}
	defer d.hw.CS.Out(gpio.High) // We must get CS pin HIGH either way

	if err := d.hw.SPI.Tx(w, r); err != nil {
		return fmt.Errorf("Could not send or read data: %w", err)
	}

	return nil
}

func (d *LoraModem) WriteRegister(address uint16, data []uint8) error {
	log := slog.With("func", "LoraModem.WriteRegister()", "params", "(uint16, []uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Write data to modem register", "address", fmt.Sprintf("0x%02X", address))

	commands := append([]uint8{
		lora.CmdWriteRegister,
		uint8(address >> 8),
		uint8(address & 0xFF),
	}, data...)

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not write data [% X] to register at address 0x%02X: %w", commands, address, err)
	}

	return nil
}

func (d *LoraModem) ReadRegister(address uint16, length uint8) ([]uint8, error) {
	log := slog.With("func", "LoraModem.ReadRegister()", "params", "(uint16, uint8)", "return", "([]uint8, error)", "package", "comm", "module", "sx1262")
	log.Debug("Read data from modem register", "address", fmt.Sprintf("0x%02X", address))

	totalLen := 1 + 2 + 1 + length // Command(1) + Address(2) + NOP(1) + Data(len)
	r := make([]uint8, totalLen)
	w := make([]uint8, totalLen)

	w[0] = lora.CmdReadRegister
	w[1] = uint8(address >> 8)
	w[2] = uint8(address & 0xFF)
	w[3] = 0x00

	if err := d.Write(w, r); err != nil {
		return nil, fmt.Errorf("Could not write data [% X] to register at address 0x%02X: %w", w, address, err)
	}

	return r[4:], nil
}

func (d *LoraModem) handleIRQ(retries uint8) {
	log := slog.With("func", "handleIRQ()", "params", "(-)", "return", "(-)", "package", "comm", "module", "sx1262")
	log.Debug("Handle LoRa IRQs")

	irq, err := d.GetIrqStatus()
	if err != nil {
		log.Warn("Could not get LoRa IRQ status; possible hardware/SPI error", "err", err)
		if err := d.ClearIrqStatus(lora.IrqAll); err != nil {
			log.Warn("Could not clear LoRa IRQ status; Using force method...", "err", err)
			d.forceClearIrq(retries)
		}

		return
	}

	if (irq & (lora.IrqCrcErr | lora.IrqHeaderErr)) > 0 {
		log.Warn("Damaged packet received; dropping it...")
		if err := d.ClearIrqStatus(lora.IrqAll); err != nil {
			log.Warn("Could not clear LoRa IRQ status; Using force method...", "err", err)
			d.forceClearIrq(retries)
		}

		return
	}

	if (irq & lora.IrqRxDone) > 0 {
		payload, err := d.ReadBuffer()

		if err != nil {
			log.Warn("Could not read LoRa RX buffer; possible hardware/SPI error", "err", err)
		} else if len(payload) > 0 {
			log.Debug("Lora data received", "data", fmt.Sprintf("% X", payload))
			select {
			case d.RxQueue <- payload: // Sent to rxChannel queue
			default:
				log.Warn("RX channel queue is full")
			}
		}
	}

	if (irq & lora.IrqTxDone) > 0 {
		if err := d.SetRx(lora.RxContinuous); err != nil {
			log.Error("Could not enable LoRa RX mode", "mode", fmt.Sprintf("% X", lora.RxContinuous), "err", err)
		}
	}

	if err := d.ClearIrqStatus(lora.IrqAll); err != nil {
		log.Warn("Could not clear LoRa IRQ status; Using force method...", "err", err)
		d.forceClearIrq(retries)
	}
}

func (d *LoraModem) forceClearIrq(retries uint8) {
	log := slog.With("func", "forceClearIrq()", "params", "(uint8)", "return", "(-)", "package", "comm", "module", "sx1262")
	log.Debug("Force IRQ cleanup")

	for i := range retries {
		err := d.ClearIrqStatus(lora.IrqAll)
		if err == nil {
			return
		}

		log.Error("Could not force clear LoRa IRQ status; possible hardware/SPI error", "attempt", i+1, "err", err)
		time.Sleep(10 * time.Millisecond)
	}

	log.Error("Could not clear LoRa IRQ status; possible hardware/SPI error")
}

func (d *LoraModem) dataTransmit(data []uint8, offset uint8, timeout uint32) {
	log := slog.With("func", "dataTransmit()", "params", "([]uint8)", "return", "(-)", "package", "comm", "module", "sx1262")
	log.Debug("Transmit data")

	if err := d.SetStandby(lora.StandbyRc); err != nil {
		log.Error("Failed to set standby mode", "err", err)
		return
	}

	if err := d.WriteBuffer(data, offset); err != nil {
		log.Error("Failed to load data to buffer", "err", err)
		return
	}

	err := d.SetPacketParams(
		d.hw.Config.PreambleLen,
		d.hw.Config.HeaderType,
		uint8(len(data)),
		d.hw.Config.CRCType,
		d.hw.Config.InvertIQ,
	)
	if err != nil {
		log.Error("Failed to set packet parameters", "err", err)
		return
	}

	if err := d.SetTx(timeout); err != nil {
		log.Error("Failed to transmit data", "err", err)
		return
	}
}
