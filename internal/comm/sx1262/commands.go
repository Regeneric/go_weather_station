package sx1262

import (
	"fmt"
	"log/slog"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
	"periph.io/x/conn/v3/physic"
)

func (d *LoraModem) SetStandby(mode uint8) error {
	log := slog.With("func", "LoraModem.SetStandby()", "params", "(uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa standby mode", "mode", fmt.Sprintf("0x%02X", mode))

	commands := []uint8{lora.CmdSetStandby, mode}
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set standby mode [% X]: %w", commands, err)
	}

	log.Info("Standby mode set", "mode", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) SetPacketType(packet uint8) error {
	log := slog.With("func", "LoraModem.SetPacketType()", "params", "(uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa packet type", "packet", fmt.Sprintf("0x%02X", packet))

	commands := []uint8{lora.CmdSetPacketType, packet}
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set packet type [% X]: %w", commands, err)
	}

	log.Info("Packet type set", "packet", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) CalibrateImage(frequencyRange []uint8) error {
	log := slog.With("func", "LoraModem.CalibrateImage()", "params", "([]uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa frequency range", "range", fmt.Sprintf("% X", frequencyRange))

	commands := append([]uint8{lora.CmdCalibrateImage}, frequencyRange...)
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set frequency range [% X]: %w", commands, err)
	}

	log.Info("Frequency range set", "range", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) SetRfFrequency(frequency physic.Frequency) error {
	log := slog.With("func", "LoraModem.SetRfFrequency()", "params", "(physic.Frequency)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa RF frequency", "frequency", fmt.Sprintf("%d Hz", frequency/physic.Hertz))

	freqHz := uint64(frequency / physic.Hertz)
	freqRf := (freqHz * 33554432) / 32000000 // Freq(Hz) * 2^25 / 32 MHz

	// Big Endian
	commands := []uint8{
		lora.CmdSetRfFrequency,
		uint8((freqRf >> 24) & 0xFF),
		uint8((freqRf >> 16) & 0xFF),
		uint8((freqRf >> 8) & 0xFF),
		uint8(freqRf & 0x0FF),
	}

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set RF frequency [% X]: %w", commands, err)
	}

	log.Info("RF frequency set", "frequency", fmt.Sprintf("%d MHz", frequency/physic.MegaHertz))
	return nil
}

func (d *LoraModem) SetPaConfig(dbm int8) error {
	log := slog.With("func", "LoraModem.SetPaConfig()", "params", "(int8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa power amplifier parameters", "dbm", dbm)

	var paDutyCycle, hpMax, device uint8
	var paLut uint8 = 0x01 // Always 0x01 for SX1261/SX1262

	if d.hw.Config.DeviceType == lora.TxPowerSX1261 {
		device = 0x01
		if dbm > lora.TxMaxPowerSX1261 {
			dbm = lora.TxMaxPowerSX1261
			log.Warn("Limiting MAX trasmit", "dbm", lora.TxMaxPowerSX1261)
		}

		if dbm < lora.TxMinPowerSX1261 {
			dbm = lora.TxMinPowerSX1261
			log.Warn("Limiting MIN trasmit power", "dbm", lora.TxMinPowerSX1261)
		}
	}
	if d.hw.Config.DeviceType == lora.TxPowerSX1262 {
		device = 0x00
		if dbm > lora.TxMaxPowerSX1262 {
			dbm = lora.TxMaxPowerSX1262
			log.Warn("Limiting MAX trasmit power", "dbm", lora.TxMaxPowerSX1262)
		}

		if dbm < lora.TxMinPowerSX1261 {
			dbm = lora.TxMinPowerSX1261
			log.Warn("Limiting MIN trasmit power", "dbm", lora.TxMinPowerSX1261)
		}
	}

	switch {
	case dbm == 22:
		paDutyCycle = 0x04
		hpMax = 0x07
	case dbm >= 20:
		paDutyCycle = 0x03
		hpMax = 0x05
	case dbm >= 17:
		paDutyCycle = 0x02
		hpMax = 0x03
	case dbm >= 14:
		paDutyCycle = 0x02
		hpMax = 0x02
	default:
		paDutyCycle = 0x02
		hpMax = 0x02
	}

	commands := []uint8{lora.CmdSetPaConfig, paDutyCycle, hpMax, device, paLut}
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set power amplifier parameters [% X]: %w", commands, err)
	}

	log.Info("Power amplifier parameters set", "dbm", dbm)
	return nil
}

func (d *LoraModem) SetTxParams(dbm int8, rampTime uint8) error {
	log := slog.With("func", "LoraModem.SetTxParams()", "params", "(int8, uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa transmit params", "dbm", dbm)

	if err := d.SetPaConfig(dbm); err != nil {
		return err
	}

	commands := []uint8{lora.CmdSetTxParams, uint8(dbm), rampTime}
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set transmit power [% X]: %w", commands, err)
	}

	log.Info("Transmit power set", "dbm", dbm)
	return nil
}

func (d *LoraModem) SetModulationParams(spreadFactor uint8, bandwidth physic.Frequency, codingRate uint8, ldro bool) error {
	log := slog.With("func", "LoraModem.SetModulationParams()", "params", "(uint8, physic.Frequency, uint8, bool)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa modulation params", "spreadFactor", spreadFactor, "bandwidth", bandwidth, "codingRate", codingRate, "ldro", ldro)

	var ld uint8 = lora.LDRO_OFF
	if ldro {
		ld = lora.LDRO_ON
	}

	var crToByte = map[uint8]uint8{
		5: lora.CR4_5,
		6: lora.CR4_6,
		7: lora.CR4_7,
		8: lora.CR4_8,
	}

	cr, ok := crToByte[codingRate]
	if !ok {
		cr = lora.CR4_5
		log.Warn("Unsupported coding rate", "codingRate", codingRate)
		log.Warn("Setting Coding Rate to 4/5")
	}

	var bwToByte = map[physic.Frequency]uint8{
		7800 * physic.Hertz:   lora.BW7800,
		10400 * physic.Hertz:  lora.BW10400,
		15600 * physic.Hertz:  lora.BW15600,
		20800 * physic.Hertz:  lora.BW20800,
		31250 * physic.Hertz:  lora.BW312000,
		41700 * physic.Hertz:  lora.BW41700,
		62500 * physic.Hertz:  lora.BW62500,
		125000 * physic.Hertz: lora.BW125000,
		250000 * physic.Hertz: lora.BW250000,
		500000 * physic.Hertz: lora.BW500000,
	}

	bw, ok := bwToByte[bandwidth]
	if !ok {
		bw = lora.BW125000
		log.Warn("Unsupported bandwidth", "bw", bandwidth)
		log.Warn("Setting bandwidth to 125 kHz")
	}

	commands := []uint8{lora.CmdSetModulationParams, spreadFactor, bw, cr, ld, 0, 0, 0, 0} // Last four bytes are unused in current LoRa implementation
	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set modulation params [% X]: %w", commands, err)
	}

	log.Info("Modulation params set", "params", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) SetPacketParams(preambleLen uint16, headerType uint8, payloadLen uint8, crcType bool, invertIQ bool) error {
	log := slog.With("func", "LoraModem.SetModulationParams()", "params", "(uint8, physic.Frequency, uint8, bool)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa packet params", "preambleLen", preambleLen, "headerType", fmt.Sprintf("% X", headerType), "payloadLen", payloadLen, "crc", crcType, "invertIQ", invertIQ)

	var crc uint8 = lora.CrcOff
	if crcType {
		crc = lora.CrcOn
	}

	var iiq uint8 = lora.IqStandard
	if invertIQ {
		iiq = lora.IqInverted
	}

	// 12-E-32-C-S
	commands := []uint8{
		lora.CmdSetPacketParams,   // 12
		uint8(preambleLen >> 8),   // MSB
		uint8(preambleLen & 0xFF), // LSB
		headerType, payloadLen,    // E-32
		crc, iiq, // C-S
		0, 0, 0,
	}

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set packet params [% X]: %w", commands, err)
	}

	log.Info("Packet params set", "params", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) SetDioIrqParams(irqMask uint16) error {
	log := slog.With("func", "LoraModem.SetDioIrqParams()", "params", "(uint16)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa DIO IRQ params", "irqMask", fmt.Sprintf("% X", irqMask))

	mask := lora.IrqTxDone | lora.IrqRxDone | lora.IrqTimeout | lora.IrqCrcErr
	commands := []uint8{
		lora.CmdSetDioIrqParams,
		uint8(mask >> 8),   // IRQ MSB
		uint8(mask & 0xFF), // IRQ LSB
		uint8(mask >> 8),   // DIO IRQ MSB
		uint8(mask & 0xFF), // DIO IRQ LSB,
		0, 0, 0, 0,
	}

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set DIO IRQ params [% X]: %w", commands, err)
	}

	log.Info("DIO IRQ params set", "params", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) SetRx(rxMode uint32) error {
	log := slog.With("func", "LoraModem.SetRx()", "params", "(uint32)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Set LoRa RX mode", "mode", fmt.Sprintf("% X", rxMode))

	commands := []uint8{
		lora.CmdSetRx,
		uint8(rxMode >> 16 & 0xFF),
		uint8(rxMode >> 8 & 0xFF),
		uint8(rxMode & 0xFF),
	}

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not set RX mode [% X]: %w", commands, err)
	}

	log.Info("RX mode set", "params", fmt.Sprintf("% X", commands))
	return nil
}

func (d *LoraModem) GetIrqStatus() (uint16, error) {
	log := slog.With("func", "LoraModem.GetIrqStatus()", "params", "([]uint8)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Get LoRa IRQ status")

	commands := []uint8{lora.CmdGetIrqStatus, 0, 0, 0}
	r := make([]uint8, len(commands))

	if err := d.Write(commands, r); err != nil {
		return 0, fmt.Errorf("Could not write data [% X]: %w", commands, err)
	}
	status := uint16(r[2])<<8 | uint16(r[3])

	log.Info("Got IRQ status", "status", fmt.Sprintf("% X", status))
	return status, nil
}

func (d *LoraModem) GetRxBufferStatus() (*RxBufferStatus, error) {
	log := slog.With("func", "LoraModem.GetRxBufferStatus()", "params", "(-)", "return", "(*RxBufferStatus, error)", "package", "comm", "module", "sx1262")
	log.Debug("Get LoRa RX buffer satus status")

	totalLen := 1 + 1 + 1 + 1 // OpCode(1) + NOP(1) + NOP(1) + NOP(1)
	r := make([]uint8, totalLen)
	w := make([]uint8, totalLen)

	w[0] = lora.CmdGetBufferStatus

	if err := d.Write(w, r); err != nil {
		return nil, fmt.Errorf("Could not write data [% X]: %w", w, err)
	}

	status := RxBufferStatus{
		PayloadLengthRx:      r[2],
		RxStartBufferPointer: r[3],
	}

	log.Info("Got RX buffer status", "status", fmt.Sprintf("% X", r[2:]))
	return &status, nil
}

func (d *LoraModem) ReadBuffer() ([]uint8, error) {
	log := slog.With("func", "LoraModem.ReadBuffer()", "params", "(uint8, uint8)", "return", "([]uint8, error)", "package", "comm", "module", "sx1262")
	log.Debug("Read LoRa buffer")

	rxBuffer, err := d.GetRxBufferStatus()
	if err != nil {
		return nil, err
	}

	offset := rxBuffer.RxStartBufferPointer
	bytes := rxBuffer.PayloadLengthRx

	totalLen := 1 + 1 + 1 + bytes // OpCode(1) + Offset(1) + NOP(1) + Data(bytes)
	r := make([]uint8, totalLen)
	w := make([]uint8, totalLen)

	w[0] = lora.CmdReadBuffer
	w[1] = offset
	w[2] = 0x00

	if err := d.Write(w, r); err != nil {
		return nil, fmt.Errorf("Could not write data [% X]: %w", w, err)
	}
	data := r[3:]

	log.Info("Read buffer successfull", "data", fmt.Sprintf("% X", data))
	return data, nil
}

func (d *LoraModem) ClearIrqStatus(irqMask uint16) error {
	log := slog.With("func", "LoraModem.ClearIrqStatus()", "params", "(uint16)", "return", "(error)", "package", "comm", "module", "sx1262")
	log.Debug("Clear IRQ status", "irq", fmt.Sprintf("% X", irqMask))

	commands := []uint8{
		lora.CmdClearIrqStatus,
		uint8(irqMask >> 8),   // MSB
		uint8(irqMask & 0xFF), // LSB
	}

	if err := d.Write(commands, nil); err != nil {
		return fmt.Errorf("Could not write data [% X]: %w", commands, err)
	}

	log.Info("IRQ status cleared")
	return nil
}
