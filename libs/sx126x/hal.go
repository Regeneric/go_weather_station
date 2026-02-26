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
	Bandwidth          uint8
	SpreadingFactor    uint8
	CodingRate         uint8
	LDRO               uint8
	Bitrate            uint64
	PulseShape         uint8
	FrequencyDeviation uint64
}

type OptionsModulation func(*ConfigModulation)

func loraCodingRate(codingRate uint8) (uint8, bool) {
	crToByte := map[uint8]uint8{
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

func fskBandwidth(bandwidth physic.Frequency) (uint8, bool) {
	bwToByte := map[physic.Frequency]uint8{
		4800 * physic.Hertz:   uint8(FskBW_4800),
		5800 * physic.Hertz:   uint8(FskBW_5800),
		7300 * physic.Hertz:   uint8(FskBW_7300),
		9700 * physic.Hertz:   uint8(FskBW_9700),
		11700 * physic.Hertz:  uint8(FskBW_11700),
		14600 * physic.Hertz:  uint8(FskBW_14600),
		19500 * physic.Hertz:  uint8(FskBW_19500),
		23400 * physic.Hertz:  uint8(FskBW_23400),
		29300 * physic.Hertz:  uint8(FskBW_29300),
		39000 * physic.Hertz:  uint8(FskBW_39000),
		46900 * physic.Hertz:  uint8(FskBW_46900),
		58600 * physic.Hertz:  uint8(FskBW_58600),
		78200 * physic.Hertz:  uint8(FskBW_78200),
		93800 * physic.Hertz:  uint8(FskBW_93800),
		117300 * physic.Hertz: uint8(FskBW_117300),
		156200 * physic.Hertz: uint8(FskBW_156200),
		187200 * physic.Hertz: uint8(FskBW_187200),
		234300 * physic.Hertz: uint8(FskBW_234300),
		312000 * physic.Hertz: uint8(FskBW_312000),
		373600 * physic.Hertz: uint8(FskBW_373600),
		467000 * physic.Hertz: uint8(FskBW_467000),
	}

	bw, ok := bwToByte[bandwidth]
	return bw, ok
}

func fskPulseShape(pulseShape float32) (uint8, bool) {
	floatToPS := map[float32]uint8{
		0.0: uint8(PulseNoFilter),
		0.3: uint8(PulseGaussianBt0_3),
		0.5: uint8(PulseGaussianBt0_5),
		0.7: uint8(PulseGaussianBt0_7),
		1.0: uint8(PulseGaussianBt1),
	}

	ps, ok := floatToPS[pulseShape]
	return ps, ok
}

func (d *Device) ModulationConfigLoRa(spreadingFactor, codingRate uint8, bandwidth physic.Frequency, ldro bool) OptionsModulation {
	log := slog.With("func", "Device.ModulationConfigLoRa()", "params", "(uint8, uint8, physic.Frequency, bool)", "return", "(OptionsModulation)", "lib", "sx126x")

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

	bw, ok := loraBandwidth(bandwidth)
	if !ok {
		bw = uint8(LoRaBW_125)
		log.Warn("Unsupported bandwidth", "bw", bandwidth)
		log.Warn("Setting bandwidth to 125 kHz")
	}

	cr, ok := loraCodingRate(codingRate)
	if !ok {
		cr = uint8(LoRaCR_4_5)
		log.Warn("Unsupported coding rate", "codingRate", codingRate)
		log.Warn("Setting Coding Rate to 4/5")
	}

	return func(cfg *ConfigModulation) {
		cfg.SpreadingFactor = sf
		cfg.Bandwidth = bw
		cfg.CodingRate = cr
		cfg.LDRO = ld
	}
}

func (d *Device) ModulationConfigFSK(bitrate, freqDeviation uint64, bandwidth physic.Frequency, pulseShape float32) OptionsModulation {
	log := slog.With("func", "Device.ModulationConfigFSK()", "params", "(uint64, uint64, physic.Frequency, float32)", "return", "(OptionsModulation)", "lib", "sx126x")

	bw, bwOk := fskBandwidth(bandwidth)
	if !bwOk {
		bw = uint8(FskBW_9700)
		log.Warn("Unsupported bandwidth in FSK mode:", "bw", bandwidth)
		log.Warn("Setting bandwidth to 9700 Hz:")
	}

	br := bitrate
	if br < FskBitrateMin || br > FskBitrateMax {
		br = 4800
		log.Warn("Unsupported bitrate:", "bitrate", bitrate)
		log.Warn("Setting bitrate to 4800 b/s:")
	}
	br = (32 * RfFrequencyXtal) / br

	ps, psOk := fskPulseShape(pulseShape)
	if !psOk {
		ps = uint8(PulseGaussianBt0_5)
		log.Warn("Unsupported Pulse Shape:", "pulseShape", pulseShape)
		log.Warn("Setting Pulse Shape to 0.5:")
	}

	fd := freqDeviation
	fd = (fd * 33554432) / RfFrequencyXtal

	return func(cfg *ConfigModulation) {
		cfg.Bandwidth = bw
		cfg.Bitrate = br
		cfg.PulseShape = ps
		cfg.FrequencyDeviation = fd
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

func (d *Device) ModulationBR(bitrate uint64) OptionsModulation {
	log := slog.With("func", "Device.ModulationBR()", "params", "(uint64)", "return", "(OptionsModulation)", "lib", "sx126x")

	br := bitrate
	if br < FskBitrateMin || br > FskBitrateMax {
		br = 4800
		log.Warn("Unsupported bitrate:", "bitrate", bitrate)
		log.Warn("Setting bitrate to 4800 b/s:")
	}
	br = (32 * RfFrequencyXtal) / br

	return func(cfg *ConfigModulation) {
		cfg.Bitrate = br
	}
}

func (d *Device) ModulationPS(pulseShape float32) OptionsModulation {
	log := slog.With("func", "Device.ModulationPS()", "params", "(float32)", "return", "(OptionsModulation)", "lib", "sx126x")

	ps, ok := fskPulseShape(pulseShape)
	if !ok {
		ps = uint8(PulseGaussianBt0_5)
		log.Warn("Unsupported Pulse Shape:", "pulseShape", pulseShape)
		log.Warn("Setting Pulse Shape to 0.5:")
	}

	return func(cfg *ConfigModulation) {
		cfg.PulseShape = ps
	}
}

func (d *Device) ModulationFD(freqDeviation uint64) OptionsModulation {
	fd := freqDeviation
	fd = (fd * 33554432) / RfFrequencyXtal

	return func(cfg *ConfigModulation) {
		cfg.FrequencyDeviation = fd
	}
}

type ConfigPacket struct {
	PreambleLength          uint16
	HeaderType              uint8
	PayloadLength           uint8
	CRC                     uint8
	IQMode                  uint8
	PreambleDetectionLength uint8
	SyncWordDetectionLength uint8
	AddresComparison        uint8
	PacketType              uint8
	CRC_FSK                 uint8
	Whitening               uint8
}

type OptionsPacket func(*ConfigPacket)

func (d *Device) PacketLoRaConfig(preambleLength uint16, headerType LoRaHeaderType, payloadLength uint8, crc LoRaCrcMode, iqMode LoRaIQMode) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PreambleLength = preambleLength
		cfg.HeaderType = uint8(headerType)
		cfg.PayloadLength = payloadLength
		cfg.CRC = uint8(crc)
		cfg.IQMode = uint8(iqMode)
	}
}

func (d *Device) PacketFskConfig(preambleDetLen FskPreambleDetector, syncWordLength FskSyncWord, addrCmp FskAddressComp, packet PacketType, crc FskCrcType, whitening FskWhitening) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PreambleDetectionLength = uint8(preambleDetLen)
		cfg.SyncWordDetectionLength = uint8(syncWordLength)
		cfg.AddresComparison = uint8(addrCmp)
		cfg.PacketType = uint8(packet)
		cfg.CRC_FSK = uint8(crc)
		cfg.Whitening = uint8(whitening)
	}
}

func (d *Device) PacketPreLen(preambleLength uint16) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PreambleLength = preambleLength
	}
}

func (d *Device) PacketHT(headerType LoRaHeaderType) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.HeaderType = uint8(headerType)
	}
}

func (d *Device) PacketPayLen(payloadLength uint8) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PayloadLength = payloadLength
	}
}

func (d *Device) PacketLoRaCRC(crc LoRaCrcMode) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.CRC = uint8(crc)
	}
}

func (d *Device) PacketFskCRC(crc FskCrcType) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.CRC_FSK = uint8(crc)
	}
}

func (d *Device) PacketIQ(iqMode LoRaIQMode) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.IQMode = uint8(iqMode)
	}
}

func (d *Device) PacketPreDet(length FskPreambleDetector) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PreambleDetectionLength = uint8(length)
	}
}

func (d *Device) PacketFskSW(length FskSyncWord) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.SyncWordDetectionLength = uint8(length)
	}
}

func (d *Device) PacketAddrCmp(mode FskAddressComp) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.AddresComparison = uint8(mode)
	}
}

func (d *Device) PacketFskType(packet PacketType) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.PacketType = uint8(packet)
	}
}

func (d *Device) PacketWhitening(whitening FskWhitening) OptionsPacket {
	return func(cfg *ConfigPacket) {
		cfg.Whitening = uint8(whitening)
	}
}

func fskPreambleDetectionLength(length uint8) (FskPreambleDetector, bool) {
	byteToDet := map[uint8]FskPreambleDetector{
		0:  PreambleDetLenOff,
		8:  PreambleDetLen8,
		16: PreambleDetLen16,
		32: PreambleDetLen32,
	}

	pd, ok := byteToDet[length]
	return pd, ok
}
func fskSyncWordDetectionLength(length uint8) (FskSyncWord, bool) {
	byteToSync := map[uint8]FskSyncWord{
		0: FskSyncWordLength0,
		1: FskSyncWordLength1,
		2: FskSyncWordLength2,
		3: FskSyncWordLength3,
		4: FskSyncWordLength4,
		5: FskSyncWordLength5,
		6: FskSyncWordLength6,
		7: FskSyncWordLength7,
		8: FskSyncWordLength8,
	}

	sd, ok := byteToSync[length]
	return sd, ok
}

func fskAddressComparison(mode uint8) (FskAddressComp, bool) {
	byteToComp := map[uint8]FskAddressComp{
		0: AddrCompOff,
		1: AddrCompNode,
		2: AddrCompAll,
	}

	ac, ok := byteToComp[mode]
	return ac, ok
}

func fskPacketType(packet string) (PacketType, bool) {
	stringToPacket := map[string]PacketType{
		"static":   PacketTypeGFSKStatic,
		"variable": PacketTypeGFSKVariable,
	}

	pt, ok := stringToPacket[packet]
	return pt, ok
}

func fskCRC(mode string) (FskCrcType, bool) {
	stringToCrc := map[string]FskCrcType{
		"0":     CRC0,
		"1":     CRC1,
		"2":     CRC2,
		"1_inv": CRC1Inv,
		"2_inv": CRC2Inv,
	}

	crc, ok := stringToCrc[mode]
	return crc, ok
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

func cadSymbolNumber(number uint8) (CadSymbolNum, bool) {
	byteToSym := map[uint8]CadSymbolNum{
		1:  CadOn1Symb,
		2:  CadOn2Symb,
		4:  CadOn4Symb,
		8:  CadOn8Symb,
		16: CadOn16Symb,
	}

	sm, ok := byteToSym[number]
	return sm, ok
}

func cadExitMode(mode uint8) (CadExitMode, bool) {
	byteToSym := map[uint8]CadExitMode{
		0: CadExitStdby,
		1: CadExitRx,
	}

	em, ok := byteToSym[mode]
	return em, ok
}

func (d *Device) EnqueueTx(payload []uint8) error {
	select {
	case d.Queue.Tx <- payload:
		return nil // All ok
	default:
		return fmt.Errorf("TX queue full - packet dropped")
	}
}

func (d *Device) DequeueRx(timeout time.Duration) ([]uint8, error) {
	select {
	case payload := <-d.Queue.Rx:
		return payload, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("RX timoeut - no data received")
	}
}

func (d *Device) WaitForIRQ(timeout time.Duration) bool {
	return d.gpio.dio.WaitForEdge(timeout)
}
