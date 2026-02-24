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
	PacketConfig(preambleLength uint16, headerType sx126x.LoRaHeaderType, payloadLength uint8, crc sx126x.LoRaCrcMode, iqMode sx126x.LoRaIQMode) sx126x.OptionsPacket
	CADConfig(symbol sx126x.CadSymbolNum, detectionPeak, detectionMin uint8, exitMode sx126x.CadExitMode, timeout uint32) sx126x.OptionsCAD

	EnqueueTx(payload []uint8) error
	DequeueRx(timeout time.Duration) ([]uint8, error)
	WaitForIRQ(timeout time.Duration) bool

	Close(sleepMode sx126x.SleepConfig) error
}

type Node struct {
	hw Transceiver
}

func Setup(modem Transceiver, cfg *sx126x.Config) (*Node, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*sx126x.Config)", "return", "(func(), error)", "package", "lora")
	log.Info("LoRa modem setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("LoRa modem disabled in the config")
	}

	if modem == nil || reflect.ValueOf(modem).IsNil() {
		return nil, func() {}, fmt.Errorf("LoRa modem state improper")
	}

	cleanup := func() {
		slog.Debug("Closing LoRa modem connection...")

		stringToSleep := map[string]sx126x.SleepConfig{
			"cold_start":     sx126x.SleepColdStart,
			"warm_start":     sx126x.SleepWarmStart,
			"cold_start_rtc": sx126x.SleepColdStartRtc,
			"warm_start_rtc": sx126x.SleepWarmStartRtc,
		}

		mode, ok := stringToSleep[cfg.SleepMode]
		if !ok {
			mode = sx126x.SleepWarmStart
			log.Warn("Unknown sleep mode", "mode", cfg.SleepMode)
			log.Warn("Limiting sleep mode to Warm Start")
		}

		if err := modem.Close(mode); err != nil {
			log.Error("Critical failure during SX126X modem shutdown", "error", err)
		}
	}

	// ************************************************************************
	// = 14.3 Circuit Configuration for Basic Rx Operation ===
	// ------------------------------------------------------------------------
	if err := modem.HardReset(); err != nil {
		return nil, func() {}, err
	}

	// = 13.1.2 SetStandby =============
	stringToStandby := map[string]sx126x.StandbyMode{
		"rc":   sx126x.StandbyRc,
		"xosc": sx126x.StandbyXosc,
	}

	standby, ok := stringToStandby[cfg.StandbyMode]
	if !ok {
		standby = sx126x.StandbyRc
		log.Warn("Unknown standby mode", "mode", cfg.StandbyMode)
		log.Warn("Limiting standby mode to RC")
	}

	if err := modem.SetStandby(standby); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.4.2 SetPacketType ==========
	stringToPacket := map[string]sx126x.PacketType{
		"lora": sx126x.PacketTypeLoRa,
		"fsk":  sx126x.PacketTypeGFSK,
	}

	packet, ok := stringToPacket[cfg.Modem]
	if !ok {
		packet = sx126x.PacketTypeLoRa
		log.Warn("Unknown packet type", "packet", cfg.Modem)
		log.Warn("Limiting pakcet type to LoRa")
	}

	if err := modem.SetPacketType(packet); err != nil {
		cleanup()
		return nil, func() {}, err
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

	freq1, freq1Ok := wordToFreq[cfg.FrequencyRange[0]]
	freq2, freq2Ok := wordToFreq[cfg.FrequencyRange[1]]

	if !freq1Ok || !freq2Ok {
		freq1 = sx126x.CalImg430
		freq2 = sx126x.CalImg440

		log.Warn("Unknown frequency", "freq1", cfg.FrequencyRange[0], "freq2", cfg.FrequencyRange[1])
		log.Warn("Limiting frequency range to 430-440")
	}

	if err := modem.CalibrateImage(freq1, freq2); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.4.1 SetRfFrequency =========
	if err := modem.SetRfFrequency(physic.Frequency(cfg.Frequency) * physic.Hertz); err != nil {
		cleanup()
		return nil, func() {}, err
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

	ramp, ok := wordToRamp[cfg.RampTime]
	if !ok {
		ramp = sx126x.PaRamp800u
		log.Warn("Unknown ramp time value", "rampTime", cfg.RampTime)
		log.Warn("Limiting ramp time to 800us")
	}

	if err := modem.SetTxParams(cfg.TransmitPower, ramp); err != nil {
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.4.8 SetBufferBaseAddress ===
	if err := modem.SetBufferBaseAddress(cfg.TxBufferAddress, cfg.RxBufferAddress); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.4.5 SetModulationParams ====
	if err := modem.SetModulationParams(modem.ModulationConfigLoRa(cfg.SpreadingFactor, cfg.CodingRate, physic.Frequency(cfg.Bandwidth)*physic.Hertz, cfg.LDRO)); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.4.6 SetPacketParams ========
	header := sx126x.HeaderExplicit
	if cfg.HeaderImplicit == true {
		header = sx126x.HeaderImplicit
	}

	crc := sx126x.CrcOff
	if cfg.CRC == true {
		crc = sx126x.CrcOn
	}

	iq := sx126x.IqStandard
	if cfg.InvertedIQ == true {
		iq = sx126x.IqInverted
	}

	if err := modem.SetPacketParams(modem.PacketConfig(cfg.PreambleLength, header, cfg.PayloadLength, crc, iq)); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.3.1 SetDioIrqParams ========
	mask := sx126x.IrqTxDone | sx126x.IrqRxDone | sx126x.IrqTimeout | sx126x.IrqCrcErr | sx126x.IrqHeaderErr // Default mask, can be changed later
	if err := modem.SetDioIrqParams(mask); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = LoRa SyncWord =================
	if _, err := modem.WriteRegister(uint16(sx126x.RegLoraSyncWordMsb), []uint8{uint8(cfg.SyncWord >> 8), uint8(cfg.SyncWord)}); err != nil {
		cleanup()
		return nil, func() {}, err
	}

	syncWord := make([]uint8, 2)
	if _, err := modem.ReadRegister(uint16(sx126x.RegLoraSyncWordMsb), syncWord); err != nil {
		cleanup()
		return nil, func() {}, err
	} else {
		log.Debug("Read register success", "syncWord", fmt.Sprintf("% X", syncWord))
	}
	// ---------------------------------

	// = 13.3.5 SetDIO2AsRfSwitchCtrl ==
	if err := modem.SetDIO2AsRfSwitchCtrl(cfg.DIO2AsRfSwitch); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// = 13.1.5 SetRx ==================
	if err := modem.SetRx(uint32(sx126x.RxContinuous)); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	// ---------------------------------

	// ------------------------------------------------------------------------

	node := &Node{
		hw: modem,
	}

	return node, cleanup, nil
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
