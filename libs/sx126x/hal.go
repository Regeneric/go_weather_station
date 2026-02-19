package sx126x

import (
	"fmt"
	"log/slog"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

func (d *Device) BusyCheck(timeout <-chan time.Time, sleep ...time.Duration) error {
	log := slog.With("func", "Device.BusyCheck()", "params", "(<-chan time.Time, ...time.Duration)", "return", "(error)", "lib", "sx126x")
	log.Debug("Check SX126x module busy status")

	interval := 10 * time.Millisecond
	if len(sleep) > 0 {
		interval = sleep[0]
		log.Debug("Sleep interval changed", "interval", interval)
	}

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout!")
		default:
			if d.gpio.busy.Read() == gpio.Low {
				log.Debug("SX126x modem ready")
				return nil
			}
			time.Sleep(interval) // Avoids busy wait in loop
		}
	}
}

func (d *Device) HardReset(timeout ...<-chan time.Time) error {
	log := slog.With("func", "Device.HardReset()", "params", "(-)", "return", "(error)", "lib", "sx126x")
	log.Debug("SX126x hard reset")

	if err := d.gpio.cs.Out(gpio.High); err != nil {
		return fmt.Errorf("Failed to set CS pin state to HIGH: %w", err)
	}
	if err := d.gpio.reset.Out(gpio.Low); err != nil {
		return fmt.Errorf("Failed to set RESET pin state to LOW: %w", err)
	}
	time.Sleep(1 * time.Millisecond)
	if err := d.gpio.reset.Out(gpio.High); err != nil {
		return fmt.Errorf("Failed to set RESET pin state to HIGH: %w", err)
	}

	wait := time.After(5 * time.Second)
	if len(timeout) > 0 {
		wait = timeout[0]
	}

	if err := d.BusyCheck(wait); err != nil {
		return fmt.Errorf("Failed to reset SX126x modem: %w", err)
	}

	log.Info("SX126x modem hard reset success")
	return nil
}

func (d *Device) Write(w []uint8, r []uint8, timeout ...<-chan time.Time) error {
	log := slog.With("func", "Device.Write()", "params", "([]uint8, []uint8, ...<-chan time.Time)", "return", "(error)", "lib", "sx126x")
	log.Debug("Send data to SX126x modem")

	wait := time.After(1 * time.Second)
	if len(timeout) > 0 {
		wait = timeout[0]
	}

	if err := d.BusyCheck(wait); err != nil {
		return fmt.Errorf("SX126x modem busy: %w", err)
	}

	if err := d.gpio.cs.Out(gpio.Low); err != nil {
		return fmt.Errorf("Failed to set CS pin state to %v: %w", gpio.Low, err)
	}
	defer d.gpio.cs.Out(gpio.High) // We must get CS pin to HIGH in the end

	if err := d.SPI.Tx(w, r); err != nil {
		return fmt.Errorf("Could not send or read data: %w", err)
	}

	return nil
}

// # 13.2.1 WriteRegister Function
func (d *Device) WriteRegister(address uint16, data []uint8) (uint8, error) {
	log := slog.With("func", "Device.WriteRegister()", "params", "(uint16, []uint8)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Allow writing a block of bytes in a data memory space starting at a specific address.")

	commands := append([]uint8{
		uint8(CmdWriteRegister),
		uint8(address >> 8),
		uint8(address),
	}, data...)
	status := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, status); err != nil {
		return 0, fmt.Errorf("Could not write data to register at address 0x%04X: %w", address, err)
	}

	log.Debug("Data write to register success", "address", fmt.Sprintf("0x%04X", address))
	return status[0], nil
}

// # 13.2.2 ReadRegister Function
func (d *Device) ReadRegister(address uint16, data []uint8) (uint8, error) {
	log := slog.With("func", "Device.ReadRegister()", "params", "(uint16, []uint8)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Allow reading a block of data starting at a given address.")

	commands := make([]uint8, len(data)+4) // OPCODE + ADDRESS_MSB + ADDRESS_LSB + NOP + LEN(DATA)
	commands[0] = uint8(CmdReadRegister)
	commands[1] = uint8(address >> 8)
	commands[2] = uint8(address)
	commands[3] = uint8(OpCodeNop)

	rx := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, rx); err != nil {
		return 0, fmt.Errorf("Could not read data from register at address 0x%04X: %w", address, err)
	}

	status := rx[0]
	copy(data, rx[4:])

	log.Debug("Data read from register success", "address", fmt.Sprintf("0x%04X", address))
	return status, nil
}

// # 13.2.3 WriteBuffer Function
func (d *Device) WriteBuffer(offset uint8, data []uint8) (uint8, error) {
	log := slog.With("func", "Device.WriteBuffer()", "params", "(uint8, []uint8)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Store data payload to be transmitted.")

	commands := append([]uint8{uint8(CmdWriteBuffer), offset}, data...)
	status := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, status); err != nil {
		return 0, fmt.Errorf("Could not write data to buffer at offset 0x%02X: %w", offset, err)
	}

	log.Debug("Data write to register success", "address", fmt.Sprintf("0x%02X", offset))
	return status[0], nil
}

// # 13.2.4 ReadBuffer Function
func (d *Device) ReadBuffer(offset uint8, data []uint8) (uint8, error) {
	log := slog.With("func", "Device.ReadBuffer()", "params", "(uint8, []uint8)", "return", "(uint8, error)", "lib", "sx126x")
	log.Debug("Read bytes of payload received starting at offset")

	commands := make([]uint8, len(data)+3)
	commands[0] = uint8(CmdReadBuffer)
	commands[1] = offset

	rx := make([]uint8, len(commands))

	if err := d.SPI.Tx(commands, rx); err != nil {
		return 0, fmt.Errorf("Could not read data from bufferr at offset 0x%02X: %w", offset, err)
	}

	status := rx[0]
	copy(data, rx[3:])

	log.Debug("Data read from register success", "address", fmt.Sprintf("0x%02X", offset))
	return status, nil
}

type ConfigPa struct {
	TxPower     int8
	PaDutyCycle uint8
	HpMax       uint8
	DeviceSel   PaConfigDeviceSel
	PaLut       uint8
}

type OptionsPa func(*ConfigPa)

func (d *Device) PaConfig(txPower int8, paDutyCycle, hpMax, paLut uint8, deviceSel PaConfigDeviceSel) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.TxPower = txPower
		cfg.PaDutyCycle = paDutyCycle
		cfg.HpMax = hpMax
		cfg.DeviceSel = deviceSel
		cfg.PaLut = paLut
	}
}

func (d *Device) PaTxPower(txPower int8) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.TxPower = txPower
	}
}

func (d *Device) PaDutyCycle(paDutyCycle uint8) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.PaDutyCycle = paDutyCycle
	}
}

func (d *Device) PaHpMax(hpMax uint8) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.HpMax = hpMax
	}
}

func (d *Device) PaDeviceSel(deviceSel PaConfigDeviceSel) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.DeviceSel = deviceSel
	}
}

func (d *Device) PaLut(paLut uint8) OptionsPa {
	return func(cfg *ConfigPa) {
		cfg.PaLut = paLut
	}
}

type ConfigModulation struct {
	SpreadingFactor uint8
	Bandwidth       uint8
	CodingRate      uint8
	LDRO            uint8
}

type OptionsModulation func(*ConfigModulation)

func loraCodingRate(codingRate uint8) (uint8, bool) {
	crToByte := map[uint8]uint8{
		4: uint8(LoRaCR_4_4),
		5: uint8(LoRaCR_4_5),
		6: uint8(LoRaCR_4_6),
		7: uint8(LoRaCR_4_7),
		8: uint8(LoRaCR_4_8),
	}

	cr, ok := crToByte[codingRate]
	return cr, ok
}

func loraBandwidth(bandwidth physic.Frequency) (uint8, bool) {
	bwToByte := map[physic.Frequency]uint8{
		7800 * physic.Hertz:   uint8(LoRaBW_7_8),
		10400 * physic.Hertz:  uint8(LoRaBW_10_4),
		15600 * physic.Hertz:  uint8(LoRaBW_15_6),
		20800 * physic.Hertz:  uint8(LoRaBW_20_8),
		31250 * physic.Hertz:  uint8(LoRaBW_31_25),
		41700 * physic.Hertz:  uint8(LoRaBW_41_7),
		62500 * physic.Hertz:  uint8(LoRaBW_62_5),
		125000 * physic.Hertz: uint8(LoRaBW_125),
		250000 * physic.Hertz: uint8(LoRaBW_250),
		500000 * physic.Hertz: uint8(LoRaBW_500),
	}

	bw, ok := bwToByte[bandwidth]
	return bw, ok
}

func (d *Device) ModulationConfig(spreadingFactor, codingRate uint8, bandwidth physic.Frequency, ldro bool) OptionsModulation {
	// TODO : ADD FSK MODULATION CONFIG
	log := slog.With("func", "Device.ModulationConfig()", "params", "(uint8, uint8, physic.Frequency, bool)", "return", "(OptionsModulation)", "lib", "sx126x")

	sf := spreadingFactor
	if sf < 5 || sf > 12 {
		sf = 7
		log.Warn("Unsupported Spreading Factor", "spreadingFactor", spreadingFactor)
		log.Warn("Setting Spreading Factor to 7")
	}

	ld := uint8(LDRO_OFF)
	if ldro {
		ld = uint8(LDRO_ON)
	}

	cr, ok := loraCodingRate(codingRate)
	if !ok {
		cr = uint8(LoRaCR_4_5)
		log.Warn("Unsupported coding rate", "codingRate", codingRate)
		log.Warn("Setting Coding Rate to 4/5")
	}
	if d.Config.Modem == "lora" && cr == 4 {
		cr = uint8(LoRaCR_4_5)
		log.Warn("Unsupported coding rate in LoRa mode", "codingRate", codingRate)
		log.Warn("Setting Coding Rate to 4/5")
	}

	bw, ok := loraBandwidth(bandwidth)
	if !ok {
		bw = uint8(LoRaBW_125)
		log.Warn("Unsupported bandwidth", "bw", bandwidth)
		log.Warn("Setting bandwidth to 125 kHz")
	}

	return func(cfg *ConfigModulation) {
		cfg.SpreadingFactor = sf
		cfg.Bandwidth = bw
		cfg.CodingRate = cr
		cfg.LDRO = ld
	}
}

func (d *Device) ModulationSF(spreadingFactor uint8) OptionsModulation {
	log := slog.With("func", "Device.ModulationSF()", "params", "(uint8)", "return", "(OptionsModulation)", "lib", "sx126x")

	sf := spreadingFactor
	if sf < 5 || sf > 12 {
		sf = 7
		log.Warn("Unsupported Spreading Factor", "spreadingFactor", spreadingFactor)
		log.Warn("Setting Spreading Factor to 7")
	}

	return func(cfg *ConfigModulation) {
		cfg.SpreadingFactor = sf
	}
}

func (d *Device) ModulationBW(bandwidth physic.Frequency) OptionsModulation {
	log := slog.With("func", "Device.ModulationBW()", "params", "(physic.Frequency)", "return", "(OptionsModulation)", "lib", "sx126x")

	bw, ok := loraBandwidth(bandwidth)
	if !ok {
		bw = uint8(LoRaBW_125)
		log.Warn("Unsupported bandwidth", "bw", bandwidth)
		log.Warn("Setting bandwidth to 125 kHz")
	}

	return func(cfg *ConfigModulation) {
		cfg.Bandwidth = bw
	}
}

func (d *Device) ModulationCR(codingRate uint8) OptionsModulation {
	log := slog.With("func", "Device.ModulationCR()", "params", "(uint8)", "return", "(OptionsModulation)", "lib", "sx126x")

	cr, ok := loraCodingRate(codingRate)
	if !ok {
		cr = uint8(LoRaCR_4_5)
		log.Warn("Unsupported coding rate", "codingRate", cr)
		log.Warn("Setting Coding Rate to 4/5")
	}
	if d.Config.Modem == "lora" && cr == 4 {
		cr = uint8(LoRaCR_4_5)
		log.Warn("Unsupported coding rate in LoRa mode", "codingRate", cr)
		log.Warn("Setting Coding Rate to 4/5")
	}

	return func(cfg *ConfigModulation) {
		cfg.CodingRate = cr
	}
}

func (d *Device) ModulationLDRO(ldro bool) OptionsModulation {
	ld := uint8(LDRO_OFF)
	if ldro {
		ld = uint8(LDRO_ON)
	}

	return func(cfg *ConfigModulation) {
		cfg.LDRO = ld
	}
}

type ConfigParams struct {
	PreambleLength uint16
	HeaderType     uint8
	PayloadLength  uint8
	CRC            uint8
	IQMode         uint8
}

type OptionsParams func(*ConfigParams)

func (d *Device) ParamsConfig(preambleLength uint16, headerType LoRaHeaderType, payloadLength uint8, crc LoRaCrcMode, iqMode LoRaIQMode) OptionsParams {
	// TODO : ADD FSK MODULATION CONFIG
	return func(cfg *ConfigParams) {
		cfg.PreambleLength = preambleLength
		cfg.HeaderType = uint8(headerType)
		cfg.PayloadLength = payloadLength
		cfg.CRC = uint8(crc)
		cfg.IQMode = uint8(iqMode)
	}
}

func (d *Device) ParamsPreLen(preambleLength uint16) OptionsParams {
	return func(cfg *ConfigParams) {
		cfg.PreambleLength = preambleLength
	}
}

func (d *Device) ParamsHT(headerType LoRaHeaderType) OptionsParams {
	return func(cfg *ConfigParams) {
		cfg.HeaderType = uint8(headerType)
	}
}

func (d *Device) ParamsPayLen(payloadLength uint8) OptionsParams {
	return func(cfg *ConfigParams) {
		cfg.PayloadLength = payloadLength
	}
}

func (d *Device) ParamsCRC(crc LoRaCrcMode) OptionsParams {
	return func(cfg *ConfigParams) {
		cfg.CRC = uint8(crc)
	}
}

func (d *Device) ParamsIQ(iqMode LoRaIQMode) OptionsParams {
	return func(cfg *ConfigParams) {
		cfg.IQMode = uint8(iqMode)
	}
}

type ConfigCAD struct {
	SymbolNumber     uint8
	DetectionPeak    uint8
	DetectionMinimum uint8
	ExitMode         uint8
	Timeout          uint32
}

type OptionsCAD func(*ConfigCAD)

// # Table 13-73: Recommended Settings for cadDetPeak and cadDetMin with 4 Symbols Detection
func SFToPeakCAD(spreadingFactor uint8) (uint8, bool) {
	sfToPeak := map[uint8]uint8{
		5:  18,
		6:  19,
		7:  20,
		8:  21,
		9:  22,
		10: 23,
		11: 24,
		12: 25,
	}

	peak, ok := sfToPeak[spreadingFactor]
	return peak, ok
}

// # Table 13-73: Recommended Settings for cadDetPeak and cadDetMin with 4 Symbols Detection
func SFToMinCAD(spreadingFactor uint8) (uint8, bool) {
	sfToMin := map[uint8]uint8{
		5:  10,
		6:  10,
		7:  10,
		8:  10,
		9:  10,
		10: 10,
		11: 10,
		12: 10,
	}

	min, ok := sfToMin[spreadingFactor]
	return min, ok
}

func (d *Device) CADConfig(symbol CadSymbolNum, detectionPeak, detectionMin uint8, exitMode CadExitMode, timeout uint32) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.SymbolNumber = uint8(symbol)
		cfg.DetectionPeak = detectionPeak
		cfg.DetectionMinimum = detectionMin
		cfg.ExitMode = uint8(exitMode)
		cfg.Timeout = timeout
	}
}

func (d *Device) CADSym(symbol CadSymbolNum) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.SymbolNumber = uint8(symbol)
	}
}

func (d *Device) CADPeak(detectionPeak uint8) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.DetectionPeak = detectionPeak
	}
}

func (d *Device) CADMin(detectionMin uint8) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.DetectionMinimum = detectionMin
	}
}

func (d *Device) CADExit(exitMode CadExitMode) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.ExitMode = uint8(exitMode)
	}
}

func (d *Device) CADTimeout(timeout uint32) OptionsCAD {
	return func(cfg *ConfigCAD) {
		cfg.Timeout = timeout
	}
}
