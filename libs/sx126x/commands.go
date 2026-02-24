package sx126x

import (
	"fmt"
	"log/slog"

	"periph.io/x/conn/v3/physic"
)

// # 13.1.1 SetSleep
func (d *Device) SetSleep(mode SleepConfig) error {
	log := slog.With("func", "Device.SetSleep()", "params", "(SleepConfig)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set the device in SLEEP mode with the lowest current consumption possible", "mode", mode)

	commands := []uint8{uint8(CmdSetSleep), uint8(mode)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set sleep mode %v to %v: %w", CmdSetStandby, mode, err)
	}

	log.Info("SX126x modem sleep mode set", "mode", mode)
	return nil
}

// # 13.1.2 SetStandby
func (d *Device) SetStandby(mode StandbyMode) error {
	log := slog.With("func", "Device.SetStandby()", "params", "(StandbyMode)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set the device in a configuration mode which is at an intermediate level of consumption", "mode", mode)

	commands := []uint8{uint8(CmdSetStandby), uint8(mode)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set standby mode %v to %v: %w", CmdSetStandby, mode, err)
	}

	log.Info("SX126x modem standby mode set", "mode", mode)
	return nil
}

// # 13.1.3 SetFs
func (d *Device) SetFs() error {
	log := slog.With("func", "Device.SetFs()", "params", "(-)", "return", "(error)", "lib", "sx126x")
	log.Debug("set the device in the frequency synthesis mode where the PLL is locked to the carrier frequency")

	if err := d.SPI.Tx([]uint8{uint8(CmdSetFs)}, nil); err != nil {
		return fmt.Errorf("Could not set frequency synthesis mode %v: %w", CmdSetFs, err)
	}

	log.Info("SX126x modem frequency synthesis mode set")
	return nil
}

// # 13.1.4 SetTx
func (d *Device) SetTx(timeout uint32) error {
	log := slog.With("func", "Device.SetTx()", "params", "(uint32)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set device in transmit mode", "timeout", timeout)

	commands := []uint8{
		uint8(CmdSetTx),
		uint8(timeout >> 16),
		uint8(timeout >> 8),
		uint8(timeout),
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set device in transmit mode %v: %w", CmdSetTx, err)
	}

	log.Info("SX126x modem set in transit mode")
	return nil
}

// # 13.1.5 SetRx
func (d *Device) SetRx(timeout uint32) error {
	log := slog.With("func", "Device.SetRx()", "params", "(uint32)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set device in receiver mode", "timeout", timeout)

	commands := []uint8{
		uint8(CmdSetRx),
		uint8(timeout >> 16),
		uint8(timeout >> 8),
		uint8(timeout),
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set device in receiver mode %v: %w", CmdSetRx, err)
	}

	log.Info("SX126x modem set in receiver mode")
	return nil
}

// # 13.1.6 StopTimerOnPreamble
func (d *Device) StopTimerOnPreamble(enable bool) error {
	log := slog.With("func", "Device.StopTimerOnPreamble()", "params", "(bool)", "return", "(error)", "lib", "sx126x")
	log.Debug("Select if the timer is stopped upon preamble detection or Sync Word / header detection", "enable", enable)

	var param uint8 = 0x00
	if enable {
		param = 0x01
	}

	commands := []uint8{uint8(CmdStopOnPreamble), param}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set timer detection param %v: %w", CmdStopOnPreamble, err)
	}

	log.Info("SX126x modem timer param set", "enable", enable)
	return nil
}

// # 13.1.7 SetRxDutyCycle
func (d *Device) SetRxDutyCycle(rxPeriod, sleepPeriod uint32) error {
	log := slog.With("func", "Device.SetRxDutyCycle()", "params", "(uint32, uint32)", "return", "(error)", "lib", "sx126x")
	log.Debug("Sets the chip in sniff mode so that it regularly looks for new packets. This is the listen mode.")

	rp := []uint8{uint8(rxPeriod >> 16), uint8(rxPeriod >> 8), uint8(rxPeriod)}
	sp := []uint8{uint8(sleepPeriod >> 16), uint8(sleepPeriod >> 8), uint8(sleepPeriod)}

	commands := []uint8{uint8(CmdSetRxDutyCycle)}
	commands = append(commands, rp...)
	commands = append(commands, sp...)

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set RX duty cycle; RX period: %v ; Sleep period: %v: %w", rxPeriod, sleepPeriod, err)
	}

	log.Info("SX126x modem sleep and RX periods set", "rx", rxPeriod, "sleep", sleepPeriod)
	return nil
}

// # 13.1.8 SetCAD
func (d *Device) SetCAD() error {
	log := slog.With("func", "Device.SetCAD()", "params", "(-)", "return", "(error)", "lib", "sx126x")
	log.Debug("LoRa specific mode of operation where the device searches for the presence of a preamble signal")

	if d.Config.Modem != "lora" {
		return fmt.Errorf("Channel Activity Detection is a LoRa specific mode of operation")
	}

	if err := d.SPI.Tx([]uint8{uint8(CmdSetCad)}, nil); err != nil {
		return fmt.Errorf("Could not set Channel Activity Detection mode %v: %w", CmdSetCad, err)
	}

	log.Info("SX126x modem Channel Activity Detection mode set")
	return nil
}

// # 13.1.9 SetTxContinuousWave
func (d *Device) SetTxContinuousWave() error {
	log := slog.With("func", "Device.SetTxContinuousWave()", "params", "(-)", "return", "(error)", "lib", "sx126x")
	log.Debug("Test command available for all packet types to generate a continuous wave (RF tone)")

	if err := d.SPI.Tx([]uint8{uint8(CmdSetTxContinuousWave)}, nil); err != nil {
		return fmt.Errorf("Could not set TX continuous wave mode %v: %w", CmdSetTxContinuousWave, err)
	}

	log.Info("SX126x modem TX continuous wave mode set")
	return nil
}

// # 13.1.10 SetTxInfinitePreamble
func (d *Device) SetTxInfinitePreamble() error {
	log := slog.With("func", "Device.SetTxInfinitePreamble()", "params", "(-)", "return", "(error)", "lib", "sx126x")
	log.Debug("FSK: Test command to generate an infinite sequence of alternating zeros and ones")
	log.Debug("LoRa: Constantly modulate LoRa preamble symbols")

	if err := d.SPI.Tx([]uint8{uint8(CmdSetTxInfinitePreamble)}, nil); err != nil {
		return fmt.Errorf("Could not set TX infinite preamble mode %v: %w", CmdSetTxInfinitePreamble, err)
	}

	log.Info("SX126x modem TX infinie preamble mode set")
	return nil
}

// # 13.1.11 SetRegulatorMode
func (d *Device) SetRegulatorMode(mode RegulatorMode) error {
	log := slog.With("func", "Device.SetRegulatorMode()", "params", "(RegulatorMode)", "return", "(error)", "lib", "sx126x")
	log.Debug("Allow to specify if DC-DC or LDO is used for power regulation", "mode", mode)

	commands := []uint8{uint8(CmdSetRegulatorMode), uint8(mode)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set power regulator mode %v: %w", mode, err)
	}

	log.Info("SX126x modem power regulator mode set")
	return nil
}

// # 13.1.12 Calibrate Function
func (d *Device) Calibrate(param CalibrationParam) error {
	log := slog.With("func", "Device.Calibrate()", "params", "(CalibrationParam)", "return", "(error)", "lib", "sx126x")
	log.Debug("Calibrate function starts the calibration of a block defined by param", "param", param)

	commands := []uint8{uint8(CmdCalibrate), uint8(param)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not calibrate modem %v: %w", param, err)
	}

	log.Info("SX126x calibration param set", "param", param)
	return nil
}

// # 13.1.13 CalibrateImage
func (d *Device) CalibrateImage(freq1, freq2 CalibrationImageFreq) error {
	log := slog.With("func", "Device.CalibrateImage()", "params", "(CalibrationImageFreq, CalibrationImageFreq)", "return", "(error)", "lib", "sx126x")
	log.Debug("Device operating frequency band")

	commands := []uint8{uint8(CmdCalibrateImage), uint8(freq1), uint8(freq2)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set modem frequency band: %v - %v: %w", freq1, freq2, err)
	}

	log.Info("SX126x modem frequency band set", "freq1", freq1, "freq2", freq2)
	return nil
}

// # 13.1.14 SetPaConfig
//
// The following restrictions must be observed to avoid voltage overstress on the PA, exceeding the maximum ratings may cause irreversible damage to the device:
//   - For SX1261 at synthesis frequency above 400 MHz, paDutyCycle should not be higher than 0x07.
//   - For SX1261 at synthesis frequency below 400 MHz, paDutyCycle should not be higher than 0x04.
//   - For SX1262, paDutyCycle should not be higher than 0x04.
func (d *Device) SetPaConfig(opts ...OptionsPa) error {
	log := slog.With("func", "Device.SetPaConfig()", "params", "(...OptionsPa)", "return", "(error)", "lib", "sx126x")
	log.Debug("Differentiate the SX1261 from the SX1262")

	cfg := &ConfigPa{TxPower: d.Config.TransmitPower, PaLut: 0x01} // PaLut should be always 0x01 for SX1261 and 1262

	switch d.Config.Type {
	case "1261":
		cfg.DeviceSel = TxPowerSX1261
		if cfg.TxPower > TxMaxPowerSX1261 {
			cfg.TxPower = TxMaxPowerSX1261
			log.Warn("Limiting MAX transmit power", "dbm", TxMaxPowerSX1261)
		}

		if cfg.TxPower < TxMinPowerSX1261 {
			cfg.TxPower = TxMinPowerSX1261
			log.Warn("Limiting MIN transmit power", "dbm", TxMinPowerSX1261)
		}
	case "1262":
		cfg.DeviceSel = TxPowerSX1262
		if cfg.TxPower > TxMaxPowerSX1262 {
			cfg.TxPower = TxMaxPowerSX1262
			log.Warn("Limiting MAX transmit power", "dbm", TxMaxPowerSX1262)
		}

		if cfg.TxPower < TxMinPowerSX1262 {
			cfg.TxPower = TxMinPowerSX1262
			log.Warn("Limiting MIN transmit power", "dbm", TxMinPowerSX1262)
		}
	default:
		return fmt.Errorf("Uknown LoRa modem type %v", d.Config.Type)
	}

	for _, opt := range opts {
		opt(cfg)
	}

	manualMode := len(opts) > 0

	// Default values for given TX power
	if !manualMode {
		switch d.Config.Type {
		case "1261":
			switch {
			case cfg.TxPower == 15:
				cfg.PaDutyCycle = 0x06
				cfg.HpMax = 0x00
			case cfg.TxPower == 14:
				cfg.PaDutyCycle = 0x04
				cfg.HpMax = 0x00
			case cfg.TxPower == 10:
				cfg.PaDutyCycle = 0x01
				cfg.HpMax = 0x00
			default:
				cfg.PaDutyCycle = 0x01
				cfg.HpMax = 0x00
			}
		case "1262":
			switch {
			case cfg.TxPower == 22:
				cfg.PaDutyCycle = 0x04
				cfg.HpMax = 0x07
			case cfg.TxPower >= 20:
				cfg.PaDutyCycle = 0x03
				cfg.HpMax = 0x05
			case cfg.TxPower >= 17:
				cfg.PaDutyCycle = 0x02
				cfg.HpMax = 0x03
			case cfg.TxPower >= 14:
				cfg.PaDutyCycle = 0x02
				cfg.HpMax = 0x02
			default:
				cfg.PaDutyCycle = 0x02
				cfg.HpMax = 0x02
			}
		default:
			return fmt.Errorf("Uknown LoRa modem type %v", d.Config.Type)
		}
	} else {
		log.Info("Manual PA config detected - skipping auto-tuning")
	}

	commands := []uint8{uint8(CmdSetPaConfig), cfg.PaDutyCycle, cfg.HpMax, uint8(cfg.DeviceSel), cfg.PaLut}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set PA calibration params: %w", err)
	}

	log.Info("SX126x modem PA calibration params set")
	return nil
}

// # 13.1.15 SetRxTxFallbackMode
func (d *Device) SetRxTxFallbackMode(mode FallbackMode) error {
	log := slog.With("func", "Device.SetRxTxFallbackMode()", "params", "(FallbackMode)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set mode the chip goes after a successful transmission or after a packet reception")

	commands := []uint8{uint8(CmdSetRxTxFallbackMode), uint8(mode)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set RX/TX fallback mode to %v: %w", mode, err)
	}

	log.Info("SX126x modem RX/TX fallback mode set", "mode", mode)
	return nil
}

// # 13.3.1 SetDioIrqParams
func (d *Device) SetDioIrqParams(irqMask IrqMask, dioIRQ ...IrqMask) error {
	log := slog.With("func", "Device.SetDioIrqParams()", "params", "(uint16, ...uint16)", "return", "(error)", "lib", "sx126x")
	log.Debug("Mask or unmask the IRQ which can be triggered by the device")

	switch d.Config.Modem {
	case "lora":
		fskBits := IrqSyncWordValid
		if irqMask&fskBits != 0 {
			return fmt.Errorf("SyncWordValid IRQ available only in FSK mode")
		}
	case "fsk":
		illegalBits := map[IrqMask]string{
			IrqHeaderValid: "IrqHeaderValid",
			IrqHeaderErr:   "IrqHeaderErr",
			IrqCadDone:     "IrqCadDone",
			IrqCadDetected: "IrqCadDetected",
		}

		for bit, name := range illegalBits {
			if irqMask&bit != 0 {
				return fmt.Errorf("%s IRQ available only in LoRa mode", name)
			}
		}
	default:
		return fmt.Errorf("Unknown modem type: %v", d.Config.Modem)
	}

	irqs := make([]uint8, 8) // IRQ mask + DIO1 + DIO2 + DIO3
	irqs[0] = uint8(irqMask >> 8)
	irqs[1] = uint8(irqMask)

	if len(dioIRQ) == 0 {
		irqs[2] = irqs[0]
		irqs[3] = irqs[1]
	} else if len(dioIRQ) > 0 && len(dioIRQ) <= 3 {
		for i, v := range dioIRQ {
			idx := 2 + (i * 2)
			irqs[idx] = uint8(v >> 8)
			irqs[idx+1] = uint8(v & 0xFF)
		}
	} else {
		return fmt.Errorf("Could not set IRQ na DIO masks; invalid number of IRQ params: %v", len(dioIRQ))
	}

	commands := append([]uint8{uint8(CmdSetDioIrqParams)}, irqs...)
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set IRQ and DIO masks: %w", err)
	}

	log.Info("SX126x modem IRQ and DIO masks set")
	return nil
}

// # 13.3.3 GetIrqStatus
func (d *Device) GetIrqStatus() (uint16, error) {
	log := slog.With("func", "Device.GetIrqStatus()", "params", "(-)", "return", "(uint16, error)", "lib", "sx126x")
	log.Debug("Returns value of the IRQ register")

	if d.Config.Modem != "lora" && d.Config.Modem != "fsk" {
		return 0, fmt.Errorf("Unknown modem type: %v", d.Config.Modem)
	}

	commands := []uint8{uint8(CmdGetIrqStatus), OpCodeNop, OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, rx); err != nil {
		return 0, fmt.Errorf("Could not get IRQ register status: %w", err)
	}
	status := uint16(rx[2])<<8 | uint16(rx[3])

	log.Info("SX126x modem IRQ register value", "status", status)
	return status, nil
}

// # 13.3.4 ClearIrqStatus
func (d *Device) ClearIrqStatus(mask IrqMask) error {
	log := slog.With("func", "Device.ClearIrqStatus()", "params", "(uint16)", "return", "(error)", "lib", "sx126x")
	log.Debug("Clear IRQ register mask")

	switch d.Config.Modem {
	case "lora":
		fskBits := IrqSyncWordValid
		if mask&fskBits != 0 {
			return fmt.Errorf("SyncWordValid IRQ available only in FSK mode")
		}
	case "fsk":
		illegalBits := map[IrqMask]string{
			IrqHeaderValid: "IrqHeaderValid",
			IrqHeaderErr:   "IrqHeaderErr",
			IrqCadDone:     "IrqCadDone",
			IrqCadDetected: "IrqCadDetected",
		}

		for bit, name := range illegalBits {
			if mask&bit != 0 {
				return fmt.Errorf("%s IRQ available only in LoRa mode", name)
			}
		}
	default:
		return fmt.Errorf("Unknown modem type: %v", d.Config.Modem)
	}

	commands := []uint8{uint8(CmdClearIrqStatus), OpCodeNop, uint8(mask >> 8), uint8(mask)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not clear IRQ register mask: %w", err)
	}

	log.Info("SX126x modem IRQ register mask cleared", "mask", mask)
	return nil
}

// # 13.3.5 SetDIO2AsRfSwitchCtrl
func (d *Device) SetDIO2AsRfSwitchCtrl(enable bool) error {
	log := slog.With("func", "Device.SetDIO2AsRfSwitchCtrl()", "params", "(bool)", "return", "(error)", "lib", "sx126x")
	log.Debug("Configure DIO2 so that it can be used to control an external RF switch")

	var extSw uint8 = 0x00
	if enable {
		extSw = 0x01
	}

	commands := []uint8{uint8(CmdSetDio2AsRfSwitchCtrl), extSw}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set DIO2 as external RF switch: %w", err)
	}

	log.Info("SX126x modem DIO2 set as external DIO switch", "enable", enable)
	return nil
}

// # 13.3.6 SetDIO3AsTCXOCtrl
func (d *Device) SetDIO3AsTCXOCtrl(voltage TcxoVoltage, timeout uint32) error {
	log := slog.With("func", "Device.SetDIO3AsTCXOCtrl()", "params", "(TcxoVoltage, uint32)", "return", "(error)", "lib", "sx126x")
	log.Debug("Configure the chip for an external TCXO reference voltage controlled by DIO3")

	commands := []uint8{
		uint8(CmdSetDio3AsTcxoCtrl),
		uint8(voltage),
		uint8(timeout >> 16),
		uint8(timeout >> 8),
		uint8(timeout),
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set DIO3 as external reference voltage %v before timeout %v: %w", voltage, timeout, err)
	}

	log.Info("SX126x modem DIO3 set as external reference voltage", "voltage", voltage)
	return nil
}

// # 13.4.1 SetRfFrequency
func (d *Device) SetRfFrequency(frequency physic.Frequency) error {
	log := slog.With("func", "Device.SetRfFrequency()", "params", "(physic.Frequency)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set the frequency of the RF frequency mode")

	freqHz := uint64(frequency / physic.Hertz)
	freqRf := (freqHz * RfFrequencyNom) / RfFrequencyXtal // Freq(Hz) * 2^25 / 32 MHz

	commands := []uint8{
		uint8(CmdSetRfFrequency),
		uint8(freqRf >> 24),
		uint8(freqRf >> 16),
		uint8(freqRf >> 8),
		uint8(freqRf),
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set RF frequency [% X]: %w", commands, err)
	}

	log.Info("SX126x modem frequency set", "frequency", fmt.Sprintf("%d MHz", frequency/physic.MegaHertz))
	return nil
}

// # 13.4.2 SetPacketType
//
// Command SetPacketType(...) must be the first of the radio configuration sequence.
func (d *Device) SetPacketType(packet PacketType) error {
	log := slog.With("func", "Device.SetPacketType()", "params", "(PacketType)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set the SX126x radio in LoRa or in FSK mode")

	commands := []uint8{uint8(CmdSetPacketType), uint8(packet)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set packet type %v: %w", packet, err)
	}

	log.Info("SX126x modem packet type set", "packet", packet)
	return nil
}

// # 13.4.3 GetPacketType
func (d *Device) GetPacketType() (uint8, error) {
	log := slog.With("func", "Device.GetPacketType()", "params", "(-)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Return current operating packet type of the radio")

	commands := []uint8{uint8(CmdGetPacketType), OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, rx); err != nil {
		return 0, fmt.Errorf("Could not get packet type: %w", err)
	}
	packet := rx[2]

	log.Info("SX126x modem packet type", "packet", packet)
	return packet, nil
}

// # 13.4.4 SetTxParams
func (d *Device) SetTxParams(dbm int8, rampTime RampTime) error {
	log := slog.With("func", "Device.GetPacketType()", "params", "(-)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Set TX output power")

	commands := []uint8{uint8(CmdSetTxParams), uint8(dbm), uint8(rampTime)}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set TX output power %v and ramp time %v: %w", dbm, rampTime, err)
	}

	log.Info("SX126x modem TX output power set", "dbm", dbm, "rampTime", rampTime)
	return nil
}

// # 13.4.5 SetModulationParams
func (d *Device) SetModulationParams(opts ...OptionsModulation) error {
	log := slog.With("func", "Device.SetModulationParams()", "params", "(...OptionsModulation)", "return", "(error)", "lib", "sx126x")
	log.Debug("Configure modulation parameters of the radio")

	cfg := ConfigModulation{}

	switch d.Config.Modem {
	case "lora":
		bw, bwOk := loraBandwidth(physic.Frequency(d.Config.Bandwidth * uint64(physic.Hertz)))
		if !bwOk {
			bw = uint8(LoRaBW_125)
			log.Warn("Unsupported bandwidth in LoRa mode", "bw", d.Config.Bandwidth)
			log.Warn("Setting bandwidth to 125 kHz")
		}

		sf := d.Config.SpreadingFactor
		if sf < 5 || sf > 12 {
			sf = 7
			log.Warn("Unsupported Spreading Factor", "spreadingFactor", d.Config.SpreadingFactor)
			log.Warn("Setting Spreading Factor to 7")
		}

		cr, crOk := loraCodingRate(d.Config.CodingRate)
		if !crOk {
			cr = uint8(LoRaCR_4_5)
			log.Warn("Unsupported Coding Rate", "codingRate", d.Config.CodingRate)
			log.Warn("Setting Coding Rate to 4/5")

		}

		ld := uint8(LDRO_OFF)
		if d.Config.LDRO {
			ld = uint8(LDRO_ON)
		}

		cfg.SpreadingFactor = sf
		cfg.Bandwidth = bw
		cfg.CodingRate = cr
		cfg.LDRO = ld
	case "fsk":
		bw, bwOk := fskBandwidth(physic.Frequency(d.Config.Bandwidth * uint64(physic.Hertz)))
		if !bwOk {
			bw = uint8(FskBW_9700)
			log.Warn("Unsupported bandwidth in FSK mode:", "bw", d.Config.Bandwidth)
			log.Warn("Setting bandwidth to 9700 Hz:")
		}

		br := d.Config.FSK.Bitrate
		if br < FskBitrateMin || br > FskBitrateMax {
			br = 4800
			log.Warn("Unsupported bitrate:", "bitrate", d.Config.FSK.Bitrate)
			log.Warn("Setting bitrate to 4800 b/s:")
		}
		br = (32 * RfFrequencyXtal) / br

		ps, psOk := fskPulseShape(d.Config.FSK.PulseShape)
		if !psOk {
			ps = uint8(PulseGaussianBt0_5)
			log.Warn("Unsupported Pulse Shape:", "pulseShape", d.Config.FSK.PulseShape)
			log.Warn("Setting Pulse Shape to 0.5:")
		}

		fd := d.Config.FSK.FrequencyDeviation
		fd = (fd * 33554432) / RfFrequencyXtal

		cfg.Bandwidth = bw
		cfg.Bitrate = br
		cfg.PulseShape = ps
		cfg.FrequencyDeviation = fd
	default:
		return fmt.Errorf("Unknown modem type: %v", d.Config.Modem)
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	var commands []uint8
	if d.Config.Modem == "lora" {
		commands = []uint8{
			uint8(CmdSetModulationParams),
			cfg.SpreadingFactor, cfg.Bandwidth,
			cfg.CodingRate, cfg.LDRO,
		}
	}
	if d.Config.Modem == "fsk" {
		commands = []uint8{
			uint8(CmdSetModulationParams),
			uint8(cfg.Bitrate >> 16),
			uint8(cfg.Bitrate >> 8),
			uint8(cfg.Bitrate),
			cfg.PulseShape, cfg.Bandwidth,
			uint8(cfg.FrequencyDeviation >> 16),
			uint8(cfg.FrequencyDeviation >> 8),
			uint8(cfg.FrequencyDeviation),
		}
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set modulation params: %w", err)
	}

	log.Info("SX126x modem modulation params set")
	return nil
}

// # 13.4.6 SetPacketParams
func (d *Device) SetPacketParams(opts ...OptionsPacket) error {
	// TODO : ADD FSK MODULATION CONFIG
	log := slog.With("func", "Device.SetPacketParams()", "params", "(...OptionsParams)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set parameters of the packet handling block")

	headerType := uint8(HeaderExplicit)
	if d.Config.HeaderImplicit {
		headerType = uint8(HeaderImplicit)
	}

	crc := uint8(CrcOff)
	if d.Config.CRC {
		crc = uint8(CrcOn)
	}

	iq := uint8(IqStandard)
	if d.Config.InvertedIQ {
		iq = uint8(IqInverted)
	}

	cfg := &ConfigPacket{
		PreambleLength: d.Config.PreambleLength,
		HeaderType:     headerType,
		PayloadLength:  d.Config.PayloadLength,
		CRC:            crc,
		IQMode:         iq,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	commands := []uint8{
		uint8(CmdSetPacketParams),
		uint8(cfg.PreambleLength >> 8),
		uint8(cfg.PreambleLength),
		cfg.HeaderType, cfg.PayloadLength,
		cfg.CRC, cfg.IQMode,
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set packet parameters: %w", err)
	}

	log.Info("SX126x modem parameters set")
	return nil
}

// # 13.4.7 SetCadParams
func (d *Device) SetCadParams(opts ...OptionsCAD) error {
	log := slog.With("func", "Device.SetCadParams()", "params", "(...OptionsCAD)", "return", "(error)", "lib", "sx126x")
	log.Debug("Define number of symbols on which CAD operates")

	cfg := &ConfigCAD{
		SymbolNumber:     d.Config.CAD.SymbolNumber,
		DetectionPeak:    d.Config.CAD.DetectionPeak,
		DetectionMinimum: d.Config.CAD.DetectionMinimum,
		ExitMode:         d.Config.CAD.ExitMode,
		Timeout:          d.Config.CAD.Timeout,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	commands := []uint8{
		uint8(CmdSetCadParams),
		cfg.SymbolNumber,
		cfg.DetectionPeak,
		cfg.DetectionMinimum,
		cfg.ExitMode,
		uint8(cfg.Timeout >> 16),
		uint8(cfg.Timeout >> 8),
		uint8(cfg.Timeout),
	}

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set CAD params: %w", err)
	}

	log.Info("SX126x modem CAD params set")
	return nil
}

// # 13.4.8 SetBufferBaseAddress
func (d *Device) SetBufferBaseAddress(txBaseAddress, rxBaseAddress uint8) error {
	log := slog.With("func", "Device.SetBufferBaseAddress()", "params", "(uint8, uint8)", "return", "(error)", "lib", "sx126x")
	log.Debug("Set the base addresses in the data buffer in all modes of operations for the packet handing operation in TX and RX mode")

	commands := []uint8{uint8(CmdSetBufferBaseAddress), txBaseAddress, rxBaseAddress}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set TX and RX buffer base addresses: %w", err)
	}

	log.Info("TX and RX base addresses set", "tx", txBaseAddress, "rx", rxBaseAddress)
	return nil
}

// # 13.4.9 SetLoRaSymbNumTimeout
func (d *Device) SetLoRaSymbNumTimeout(symbols uint8) error {
	log := slog.With("func", "Device.SetLoRaSymbNumTimeout()", "params", "(uint8)", "return", "(error)", "lib", "sx126x")
	log.Debug("When the `symbols` param is set the 0, the modem will validate the reception as soon as a LoRa Symbol has been detected")

	commands := []uint8{uint8(CmdSetSymbNumTimeout), symbols}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not set symbols number timeout: %w", err)
	}

	log.Info("SX126x modem symbols numbers timeout set", "symbols", symbols)
	return nil
}

// # 13.5.1 GetStatus
func (d *Device) GetStatus() (ModemStatus, error) {
	log := slog.With("func", "Device.GetStatus()", "params", "(-)", "return", "(ModemStatus, error)", "lib", "sx126x")
	log.Debug("Retrieve chip status directly")

	tx := []uint8{uint8(CmdGetStatus), OpCodeNop}
	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return ModemStatus{}, fmt.Errorf("Could not get modem status: %w", err)
	}

	status := rx[1]
	d.Status.Modem.ChipMode = (status & 0x70) >> 4 // Bits 6:4
	d.Status.Modem.Command = (status & 0x0E) >> 1  // Bits 3:1

	log.Info("SX126x modem status", "command", d.Status.Modem.Command, "chip", d.Status.Modem.ChipMode)
	return d.Status.Modem, nil
}

// # 13.5.2 GetRxBufferStatus
func (d *Device) GetRxBufferStatus() (BufferStatus, error) {
	log := slog.With("func", "Device.GetRxBufferStatus()", "params", "(-)", "return", "(BufferStatus, error)", "lib", "sx126x")
	log.Debug("Return the length of the last received packet and the address of the first byte received")

	tx := []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return BufferStatus{}, fmt.Errorf("Could not get RX buffer status: %w", err)
	}

	d.Status.Buffer.RXPayloadLength = rx[2]
	d.Status.Buffer.RXStartPointer = rx[3]

	log.Info("SX126x modem RX buffer status")
	return d.Status.Buffer, nil
}

// # 13.5.3 GetPacketStatus
func (d *Device) GetPacketStatus() (PacketStatus, error) {
	log := slog.With("func", "Device.GetPacketStatus()", "params", "(-)", "return", "(PacketStatus, error)", "lib", "sx126x")
	log.Debug("Table 13-81: Status Bit")

	tx := []uint8{uint8(CmdGetPacketStatus), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return PacketStatus{}, fmt.Errorf("Could not get packet status: %w", err)
	}

	d.Status.Packet.SignalStrength = -int8(rx[2] / 2)         // dBm
	d.Status.Packet.SnRRatio = float32(int8(rx[3])) / 4.0     // dBm
	d.Status.Packet.DenoisedSignalStrength = -int8(rx[4] / 2) // dBm

	log.Info("SX126x modem packet status")
	return d.Status.Packet, nil
}

// # 13.5.4 GetRssiInst
func (d *Device) GetRssiInst() (int8, error) {
	log := slog.With("func", "Device.GetRssiInst()", "params", "(-)", "return", "(int8, error)", "lib", "sx126x")
	log.Debug("Return instantaneous RSSI value during reception of the packet")

	tx := []uint8{uint8(CmdGetPacketRssi), OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return 0, fmt.Errorf("Could not get packet instant RSSI value: %w", err)
	}
	rssi := -int8(rx[2] / 2) // dBm

	log.Info("SX126x modem packet instant RSSI value", "rssi", rssi)
	return rssi, nil
}

// # 13.5.5 GetStats
func (d *Device) GetStats() (PacketStats, error) {
	log := slog.With("func", "Device.GetStats()", "params", "(-)", "return", "(PacketStats, error)", "lib", "sx126x")
	log.Debug("Return the number of informations received on a few last packets")

	tx := make([]uint8, 8)
	tx[0] = uint8(CmdGetStats)

	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return PacketStats{}, fmt.Errorf("Could not get packet statistics: %w", err)
	}

	d.Status.Packet.Stats.TotalReceived = (uint16(rx[2])<<8 | uint16(rx[3]))
	d.Status.Packet.Stats.CrcErrors = (uint16(rx[4])<<8 | uint16(rx[5]))
	d.Status.Packet.Stats.HeaderErrors = (uint16(rx[6])<<8 | uint16(rx[7]))

	log.Info("SX126x modem packet statistics")
	return d.Status.Packet.Stats, nil
}

// # 13.5.6 ResetStats
func (d *Device) ResetStats(resetInternalCache bool) error {
	log := slog.With("func", "Device.ResetStats()", "params", "(bool)", "return", "(error)", "lib", "sx126x")
	log.Debug("Reset value read by the command GetStats")

	commands := make([]uint8, 7) // Seven NOPs
	commands[0] = uint8(CmdResetStats)

	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not reset device statistics: %w", err)
	}

	if resetInternalCache {
		d.Status.Packet.Stats = PacketStats{}
		log.Debug("Internal stats cache cleared")
	}

	log.Info("SX126x modem reset statistics")
	return nil
}

// # 13.6.1 GetDeviceErrors
func (d *Device) GetDeviceErrors() (DeviceError, error) {
	log := slog.With("func", "Device.GetDeviceErrors()", "params", "(-)", "return", "(ModemErrors, error)", "lib", "sx126x")
	log.Debug("Return possible errors flag that could occur during different chip operation")

	tx := []uint8{uint8(CmdGetDeviceErrors), OpCodeNop, OpCodeNop, OpCodeNop}
	rx := make([]uint8, len(tx))

	if err := d.SPI.Tx(tx, rx); err != nil {
		return 0, fmt.Errorf("Could not get device errors: %w", err)
	}
	d.Status.Error = DeviceError(uint16(rx[2])<<8 | uint16(rx[3]))

	log.Info("SX126x modem device errors")
	return d.Status.Error, nil
}

// # 13.6.2 ClearDeviceErrors
func (d *Device) ClearDeviceErrors(resetInternalCache bool) error {
	log := slog.With("func", "Device.ClearDeviceErrors()", "params", "(bool)", "return", "(error)", "lib", "sx126x")
	log.Debug("Clear all errors recorded in the device")

	commands := []uint8{uint8(CmdResetErrors), OpCodeNop}
	if err := d.SPI.Tx(commands, nil); err != nil {
		return fmt.Errorf("Could not reset device errors: %w", err)
	}

	if resetInternalCache {
		d.Status.Error = 0
		log.Debug("Internal errors cache cleared")
	}

	log.Info("SX126x modem reset errors")
	return nil
}
