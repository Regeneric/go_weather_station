package lora

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sx126x"
	"time"

	"periph.io/x/conn/v3/physic"
)

type Transceiver interface {
	SetSleep(mode sx126x.SleepConfig) error
	SetStandby(mode sx126x.StandbyMode) error
	SetFs() error
	SetTx(timeout uint32) error
	SetRx(timeout uint32) error
	StopTimerOnPreamble(enable bool) error
	SetRxDutyCycle(rxPeriod, sleepPeriod uint32) error
	SetCAD() error
	SetTxContinuousWave() error
	SetTxInfinitePreamble() error
	SetRegulatorMode(mode sx126x.RegulatorMode) error
	Calibrate(param sx126x.CalibrationParam) error
	CalibrateImage(freq1, freq2 sx126x.CalibrationImageFreq) error
	SetPaConfig(opts ...sx126x.OptionsPa) error
	SetRxTxFallbackMode(mode sx126x.FallbackMode) error
	SetDioIrqParams(irqMask sx126x.IrqMask, dioIRQ ...sx126x.IrqMask) error
	GetIrqStatus() (uint16, error)
	ClearIrqStatus(mask sx126x.IrqMask) error
	SetDIO2AsRfSwitchCtrl(enable bool) error
	SetDIO3AsTCXOCtrl(voltage sx126x.TcxoVoltage, timeout uint32) error
	SetRfFrequency(frequency physic.Frequency) error
	SetPacketType(packet sx126x.PacketType) error
	GetPacketType() (uint8, error)
	SetTxParams(dbm int8, rampTime sx126x.RampTime) error
	SetModulationParams(opts ...sx126x.OptionsModulation) error
	SetPacketParams(opts ...sx126x.OptionsPacket) error
	SetCadParams(opts ...sx126x.OptionsCAD) error
	SetBufferBaseAddress(txBaseAddress, rxBaseAddress uint8) error
	SetLoRaSymbNumTimeout(symbols uint8) error
	GetStatus() (sx126x.ModemStatus, error)
	GetRxBufferStatus() (sx126x.BufferStatus, error)
	GetPacketStatus() (sx126x.PacketStatus, error)
	GetRssiInst() (int8, error)
	GetStats() (sx126x.PacketStats, error)
	ResetStats(resetInternalCache bool) error
	GetDeviceErrors() (sx126x.DeviceError, error)
	ClearDeviceErrors(resetInternalCache bool) error

	BusyCheck(timeout <-chan time.Time, sleep ...time.Duration) error
	HardReset(timeout ...<-chan time.Time) error
	Write(w []uint8, r []uint8, timeout ...<-chan time.Time) error
	WriteRegister(address uint16, data []uint8) (uint8, error)
	ReadRegister(address uint16, data []uint8) (uint8, error)
	WriteBuffer(offset uint8, data []uint8) (uint8, error)
	ReadBuffer(offset uint8, data []uint8) (uint8, error)

	PaConfig(txPower int8, paDutyCycle, hpMax, paLut uint8, deviceSel sx126x.PaConfigDeviceSel) sx126x.OptionsPa
	ModulationConfigLoRa(spreadingFactor, codingRate uint8, bandwidth physic.Frequency, ldro bool) sx126x.OptionsModulation
	PacketLoRaConfig(preambleLength uint16, headerType sx126x.LoRaHeaderType, payloadLength uint8, crc sx126x.LoRaCrcMode, iqMode sx126x.LoRaIQMode) sx126x.OptionsPacket
	CADConfig(symbol sx126x.CadSymbolNum, detectionPeak, detectionMin uint8, exitMode sx126x.CadExitMode, timeout uint32) sx126x.OptionsCAD

	EnqueueTx(payload []uint8) error
	DequeueRx(timeout time.Duration) ([]uint8, error)
	WaitForIRQ(timeout time.Duration) bool

	Close(sleepMode sx126x.SleepConfig) error
}

type Node struct {
	hw  Transceiver
	cfg *sx126x.Config
}

type Status uint16

const (
	SX126XModemError Status = 0x0001
	LoraModemError   Status = 0x0002
	LoraSetupError   Status = 0x0003
)

func New(modem Transceiver, cfg *sx126x.Config) (*Node, error) {
	log := slog.With("func", "New()", "params", "(Transceiver, *sx126x.Config)", "return", "(*Node, error)", "package", "lora")
	log.Info("LoRa modem constructor")

	if cfg.Enable == false {
		return nil, fmt.Errorf("LoRa modem disabled in the config")
	}

	if modem == nil || reflect.ValueOf(modem).IsNil() {
		return nil, fmt.Errorf("LoRa modem state improper")
	}

	return &Node{
		hw:  modem,
		cfg: cfg,
	}, nil
}

func Setup(n *Node) error {
	log := slog.With("func", "Setup()", "params", "(*Node)", "return", "(error)", "package", "lora")
	log.Info("LoRa modem setup")

	if n == nil {
		return fmt.Errorf("LoRa modem state improper; node is nil")
	}
	if n.cfg == nil {
		return fmt.Errorf("LoRa modem state improper; cfg is nil")
	}
	if n.hw == nil || reflect.ValueOf(n.hw).IsNil() {
		return fmt.Errorf("LoRa modem state improper; hw is nil")
	}

	if n.cfg.Enable == false {
		return fmt.Errorf("LoRa modem disabled in the config")
	}

	// ************************************************************************
	// = 14.3 Circuit Configuration for Basic Rx Operation ===
	// ------------------------------------------------------------------------
	if err := n.hw.HardReset(); err != nil {
		return err
	}

	// = 13.1.2 SetStandby =============
	stringToStandby := map[string]sx126x.StandbyMode{
		"rc":   sx126x.StandbyRc,
		"xosc": sx126x.StandbyXosc,
	}

	standby, ok := stringToStandby[n.cfg.StandbyMode]
	if !ok {
		standby = sx126x.StandbyRc
		log.Warn("Unknown standby mode", "mode", n.cfg.StandbyMode)
		log.Warn("Limiting standby mode to RC")
	}

	if err := n.hw.SetStandby(standby); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.2 SetPacketType ==========
	stringToPacket := map[string]sx126x.PacketType{
		"lora": sx126x.PacketTypeLoRa,
		"fsk":  sx126x.PacketTypeGFSK,
	}

	packet, ok := stringToPacket[n.cfg.Modem]
	if !ok {
		packet = sx126x.PacketTypeLoRa
		log.Warn("Unknown packet type", "packet", n.cfg.Modem)
		log.Warn("Limiting pakcet type to LoRa")
	}

	if err := n.hw.SetPacketType(packet); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.1.13 CalibrateImage ========
	wordToFreq := map[uint16]sx126x.CalibrationImageFreq{
		430: sx126x.CalImg430,
		440: sx126x.CalImg440,
		470: sx126x.CalImg470,
		510: sx126x.CalImg510,
		779: sx126x.CalImg779,
		787: sx126x.CalImg787,
		863: sx126x.CalImg863,
		870: sx126x.CalImg870,
		902: sx126x.CalImg902,
		928: sx126x.CalImg928,
	}

	freq1, freq1Ok := wordToFreq[n.cfg.FrequencyRange[0]]
	freq2, freq2Ok := wordToFreq[n.cfg.FrequencyRange[1]]

	if !freq1Ok || !freq2Ok {
		freq1 = sx126x.CalImg430
		freq2 = sx126x.CalImg440

		log.Warn("Unknown frequency", "freq1", n.cfg.FrequencyRange[0], "freq2", n.cfg.FrequencyRange[1])
		log.Warn("Limiting frequency range to 430-440")
	}

	if err := n.hw.CalibrateImage(freq1, freq2); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.1 SetRfFrequency =========
	if err := n.hw.SetRfFrequency(physic.Frequency(n.cfg.Frequency) * physic.Hertz); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.4 SetTxParams ============
	wordToRamp := map[uint16]sx126x.RampTime{
		10:   sx126x.PaRamp10u,
		20:   sx126x.PaRamp20u,
		40:   sx126x.PaRamp40u,
		80:   sx126x.PaRamp80u,
		200:  sx126x.PaRamp200u,
		800:  sx126x.PaRamp800u,
		1700: sx126x.PaRamp1700u,
		3400: sx126x.PaRamp3400u,
	}

	ramp, ok := wordToRamp[n.cfg.RampTime]
	if !ok {
		ramp = sx126x.PaRamp800u
		log.Warn("Unknown ramp time value", "rampTime", n.cfg.RampTime)
		log.Warn("Limiting ramp time to 800us")
	}

	if err := n.hw.SetTxParams(n.cfg.TransmitPower, ramp); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.8 SetBufferBaseAddress ===
	if err := n.hw.SetBufferBaseAddress(n.cfg.TxBufferAddress, n.cfg.RxBufferAddress); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.5 SetModulationParams ====
	if err := n.hw.SetModulationParams(n.hw.ModulationConfigLoRa(n.cfg.LoRa.SpreadingFactor, n.cfg.LoRa.CodingRate, physic.Frequency(n.cfg.Bandwidth)*physic.Hertz, n.cfg.LoRa.LDRO)); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.4.6 SetPacketParams ========
	header := sx126x.HeaderExplicit
	if n.cfg.LoRa.HeaderImplicit == true {
		header = sx126x.HeaderImplicit
	}

	crc := sx126x.CrcOff
	if n.cfg.LoRa.CRC == true {
		crc = sx126x.CrcOn
	}

	iq := sx126x.IqStandard
	if n.cfg.LoRa.InvertedIQ == true {
		iq = sx126x.IqInverted
	}

	if err := n.hw.SetPacketParams(n.hw.PacketLoRaConfig(n.cfg.PreambleLength, header, n.cfg.PayloadLength, crc, iq)); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.3.1 SetDioIrqParams ========
	mask := sx126x.IrqTxDone | sx126x.IrqRxDone | sx126x.IrqTimeout | sx126x.IrqCrcErr | sx126x.IrqHeaderErr // Default mask, can be changed later
	if err := n.hw.SetDioIrqParams(mask); err != nil {
		return err
	}
	// ---------------------------------

	// = LoRa SyncWord =================
	if _, err := n.hw.WriteRegister(uint16(sx126x.RegLoraSyncWordMsb), []uint8{uint8(n.cfg.LoRa.SyncWord >> 8), uint8(n.cfg.LoRa.SyncWord)}); err != nil {
		return err
	}

	syncWord := make([]uint8, 2)
	if _, err := n.hw.ReadRegister(uint16(sx126x.RegLoraSyncWordMsb), syncWord); err != nil {
		return err
	} else {
		log.Debug("Read register success", "syncWord", fmt.Sprintf("% X", syncWord))
	}
	// ---------------------------------

	// = 13.3.5 SetDIO2AsRfSwitchCtrl ==
	if err := n.hw.SetDIO2AsRfSwitchCtrl(n.cfg.DIO2AsRfSwitch); err != nil {
		return err
	}
	// ---------------------------------

	// = 13.1.5 SetRx ==================
	if err := n.hw.SetRx(uint32(sx126x.RxContinuous)); err != nil {
		return err
	}
	// ---------------------------------

	// ------------------------------------------------------------------------
	return nil
}

func (n *Node) Close() error {
	log := slog.With("func", "Close()", "params", "(*Node)", "return", "(error)", "package", "lora")
	log.Info("LoRa modem destructor")

	stringToSleep := map[string]sx126x.SleepConfig{
		"cold_start":     sx126x.SleepColdStart,
		"warm_start":     sx126x.SleepWarmStart,
		"cold_start_rtc": sx126x.SleepColdStartRtc,
		"warm_start_rtc": sx126x.SleepWarmStartRtc,
	}

	mode, ok := stringToSleep[n.cfg.SleepMode]
	if !ok {
		mode = sx126x.SleepWarmStart
		log.Warn("Unknown sleep mode", "mode", n.cfg.SleepMode)
		log.Warn("Limiting sleep mode to Warm Start")
	}

	if err := n.hw.Close(mode); err != nil {
		return fmt.Errorf("Critical failure during SX126X modem shutdown: %w", err)
	}

	return nil
}

func (n *Node) Tx(data []uint8) error {
	log := slog.With("func", "Tx()", "params", "([]uint8)", "return", "(error)", "package", "lora")
	log.Info("SX126X data transmit")

	return n.hw.EnqueueTx(data)
}

func (n *Node) Run(ctx context.Context) {
	log := slog.With("func", "Run()", "params", "(context.Context)", "return", "(-)", "package", "lora")
	log.Info("SX126X modem event loop")

}
