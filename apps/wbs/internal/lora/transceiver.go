package lora

import (
	"context"
	"time"

	"github.com/Regeneric/iot-drivers/libs/sx126x"
)

type Transceiver interface {
	SetSleep(mode sx126x.SleepConfig) error
	SetStandby(mode sx126x.StandbyMode) error
	SetFs() error
	SetTx(timeout int32) error
	SetRx(timeout int32) error
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
	SetDIO3AsTCXOCtrl(voltage sx126x.TcxoVoltage, timeout int32) error
	SetRfFrequency(frequency sx126x.Frequency) error
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
	ModulationConfigLoRa(spreadingFactor, codingRate uint8, bandwidth sx126x.Frequency, ldro bool) sx126x.OptionsModulation
	PacketLoRaConfig(preambleLength uint16, headerType sx126x.LoRaHeaderType, payloadLength int, crc sx126x.LoRaCrcMode, iqMode sx126x.LoRaIQMode) sx126x.OptionsPacket
	CADConfig(symbol sx126x.CadSymbolNum, detectionPeak, detectionMin uint8, exitMode sx126x.CadExitMode, timeout uint32) sx126x.OptionsCAD

	EnqueueTx(payload []uint8) error
	DequeueRx(timeout time.Duration) ([]uint8, error)
	WaitForIRQ(timeout time.Duration) bool

	Run(ctx context.Context) error
	Close(sleepMode sx126x.SleepConfig) error
}
