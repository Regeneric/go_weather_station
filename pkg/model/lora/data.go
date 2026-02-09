package lora

import (
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type Params int

const (
	// SX126X register map
	RegFskWhiteningInitialMsb = 0x06B8
	RegFskCrcInitialMsb       = 0x06BC
	RegFskSyncWord0           = 0x06C0
	RegFskNodeAddress         = 0x06CD
	RegIqPolaritySetup        = 0x0736
	RegLoraSyncWordMsb        = 0x0740
	RegTxModulation           = 0x0889
	RegRxGain                 = 0x08AC
	RegTxClampConfig          = 0x08D8
	RegOcpConfiguration       = 0x08E7
	RegRtcControl             = 0x0902
	RegXtaTrim                = 0x0911
	RegXtbTrim                = 0x0912
	RegEventMask              = 0x0944

	// SetSleep
	SleepColdStart    = 0x00 // Cold start, configuration is lost (default)
	SleepWarmStart    = 0x04 // Warm start, configuration is retained
	SleepColdStartRtc = 0x01 // Cold start and wake on RTC timeout
	SleepWarmStartRtc = 0x05 // Warm start and wake on RTC timeout

	// SetStandby
	StandbyRc   = 0x00 // 13 MHz RC oscillator
	StandbyXosc = 0x01 // 32 MHz crystal oscillator

	// SetTx and SetRx
	TxSingle     = 0x000000 // No timeout (Tx single mode)
	RxSingle     = 0x000000 // No timeout (Rx single mode)
	RxContinuous = 0xFFFFFF // Infinite (Rx continuous mode)
	RxNoTimeout  = -1

	// SetRegulatorMode
	RegulatorLdo  = 0x00 // LOD (default)
	RegulatorDcDc = 0x01 // DC-DC

	// ISM band (MHz)
	CalImg430 = 0x6B // 430 - 440 MHz
	CalImg440 = 0x6F
	CalImg470 = 0x75 // 470 - 510 MHz
	CalImg510 = 0x81
	CalImg779 = 0xC1 // 779 - 787 MHz
	CalImg787 = 0xC5
	CalImg863 = 0xD7 // 863 - 870 MHz
	CalImg870 = 0xDB
	CalImg902 = 0xE1 // 902 - 928 MHz
	CalImg928 = 0xE9

	// SetPaConfig
	TxPowerSX1261    = 0x01
	TxPowerSX1262    = 0x02
	TxPowerSX1268    = 0x08
	TxMaxPowerSX1261 = 14
	TxMinPowerSX1261 = -17
	TxMaxPowerSX1262 = 22
	TxMinPowerSX1262 = -9

	// SetRxTxFallbackMode
	FallbackFs        = 0x40 // FS mode
	FallbackStdbyXosc = 0x30 // Crystal oscillator
	FallbackStdbyRc   = 0x20 // RC oscillator (default)

	// SetDioIrqParams
	IrqTxDone           = 0x0001 // Packet transmission completed
	IrqRxDone           = 0x0002 // Packet received
	IrqPreambleDetected = 0x0004 // Preamble detected
	IrqSyncWordValid    = 0x0008 // Valid sync word detected
	IrqHeaderValid      = 0x0010 // Valid LoRa header received
	IrqHeaderErr        = 0x0020 // LoRa header CRC error
	IrqCrcErr           = 0x0040 // Wrong CRC received
	IrqCadDone          = 0x0080 // Channel activity detection finished
	IrqCadDetected      = 0x0100 // Channel activity detected
	IrqTimeout          = 0x0200 // Rx or Tx timeout
	IrqAll              = 0x03FF // All interupts
	IrqNone             = 0x0000 // No interupts

	// SetDio2AsRfSwitch
	Dio2AsIrq      = 0x00 // IRQ
	Dio2AsRfSwitch = 0x01 // RF switch control

	// SetDio3AsTcxoCtrl
	Dio3Output1_6 = 0x00   // TCXO voltage 1.6V
	Dio3Output1_7 = 0x01   // TCXO voltage 1.7V
	Dio3Output1_8 = 0x02   // TCXO voltage 1.8V
	Dio3Output2_2 = 0x03   // TCXO voltage 2.2V
	Dio3Output2_4 = 0x04   // TCXO voltage 2.4V
	Dio3Output2_7 = 0x05   // TCXO voltage 2.7V
	Dio3Output3_0 = 0x06   // TCXO voltage 3.0V
	Dio3Output3_3 = 0x07   // TCXO voltage 3.3V
	TcxoDelay2_5  = 0x0140 // TCXO delay time 2.5 ms
	TcxoDelay5    = 0x0280 // TCXO delay time 5 ms
	TcxoDelay10   = 0x0560 // TCXO delay time 10 ms

	// SetRfFrequency
	RfFrequencyXtal = 32000000 // XTAL frequency used for RF frequency calculation
	RfFrequencyNom  = 33554432 // Used for RF frequency calculation

	// SetPacketType
	FskModem  = 0x00 // GFSK packet type
	LoraModem = 0x01 // LoRa packet type

	// SetTxParams
	PaRamp10u   = 0x00 // Ramp time 10 us
	PaRamp20u   = 0x01 // Ramp time 20 us
	PaRamp40u   = 0x02 // Ramp time 40 us
	PaRamp80u   = 0x03 // Ramp time 80 us
	PaRamp200u  = 0x04 // Ramp time 200 us
	PaRamp800u  = 0x05 // Ramp time 800 us
	PaRamp1700u = 0x06 // Ramp time 1700 us
	PaRamp3400u = 0x07 // Ramp time 3400 us

	// SetModulationParams
	BW7800   = 0x00 // LoRa bandwidth 7.8 kHz
	BW10400  = 0x08 // LoRa bandwidth 10.4 kHz
	BW15600  = 0x01 // LoRa bandwidth 15.6 kHz
	BW20800  = 0x09 // LoRa bandwidth 20.8 kHz
	BW31250  = 0x02 // LoRa bandwidth 31.25 kHz
	BW41700  = 0x0A // LoRa bandwidth 41.7 kHz
	BW62500  = 0x03 // LoRa bandwidth 62.5 kHz
	BW125000 = 0x04 // LoRa bandwidth 125.0 kHz
	BW250000 = 0x05 // LoRa bandwidth 250.0 kHz
	BW500000 = 0x06 // LoRa bandwidth 500.0 kHz
	CR4_4    = 0x00 // LoRa coding rate 4/4 (no coding rate)
	CR4_5    = 0x01 // LoRa coding rate 4/5
	CR4_6    = 0x02 // LoRa coding rate 4/6
	CR4_7    = 0x03 // LoRa coding rate 4/7
	CR4_8    = 0x04 // LoRa coding rate 4/8
	LDRO_OFF = 0x00 // LoRa low data rate optimization: disabled
	LDRO_ON  = 0x01 // LoRa low data rate optimization: enabled

	// SetModulationParams for FSK packet type
	PulseNoFilter      = 0x00 // FSK pulse shape: no filter
	PulseGaussianBt0_3 = 0x08 // FSK pulse shape: Gaussian BT 0.3
	PulseGaussianBt0_5 = 0x09 // FSK pulse shape: Gaussian BT 0.5
	PulseGaussianBt0_7 = 0x0A // FSK pulse shape: Gaussian BT 0.7
	PulseGaussianBt1   = 0x0B // FSK pulse shape: Gaussian BT 1.0
	BW4800             = 0x1F // FSK bandwidth: 4.8 kHz DSB
	BW5800             = 0x17 // FSK bandwidth: 5.8 kHz DSB
	BW7300             = 0x0F // FSK bandwidth: 7.3 kHz DSB
	BW9700             = 0x1E // FSK bandwidth: 9.7 kHz DSB
	BW11700            = 0x16 // FSK bandwidth: 11.7 kHz DSB
	BW14600            = 0x0E // FSK bandwidth: 14.6 kHz DSB
	BW19500            = 0x1D // FSK bandwidth: 19.5 kHz DSB
	BW23400            = 0x15 // FSK bandwidth: 23.4 kHz DSB
	BW29300            = 0x0D // FSK bandwidth: 29.3 kHz DSB
	BW39000            = 0x1C // FSK bandwidth: 39 kHz DSB
	BW46900            = 0x14 // FSK bandwidth: 46.9 kHz DSB
	BW58600            = 0x0C // FSK bandwidth: 58.6 kHz DSB
	BW78200            = 0x1B // FSK bandwidth: 78.2 kHz DSB
	BW93800            = 0x13 // FSK bandwidth: 93.8 kHz DSB
	BW117300           = 0x0B // FSK bandwidth: 117.3 kHz DSB
	BW156200           = 0x1A // FSK bandwidth: 156.2 kHz DSB
	BW187200           = 0x12 // FSK bandwidth: 187.2 kHz DSB
	BW234300           = 0x0A // FSK bandwidth: 234.3 kHz DSB
	BW312000           = 0x19 // FSK bandwidth: 312 kHz DSB
	BW373600           = 0x11 // FSK bandwidth: 373.3 kHz DSB
	BW467000           = 0x09 // FSK bandwidth: 476 kHz DSB

	// SetPacketParams
	HeaderExplicit = 0x00 // LoRa header mode: explicit
	HeaderImplicit = 0x01 // LoRa header mode: implicit
	CrcOff         = 0x00 // LoRa CRC mode: disabled
	CrcOn          = 0x01 // LoRa CRC mode: enabled
	IqStandard     = 0x00 // LoRa IQ setup: standard
	IqInverted     = 0x01 // LoRa IQ setup: inverted

	// SetPacketParams for FSK packet type
	PreambleDetLenOff = 0x00 // FSK preabmle detector length: off
	PreambleDetLen8   = 0x04 // FSK preabmle detector length: 8-bit
	PreambleDetLen16  = 0x05 // FSK preabmle detector length: 16-bit
	PreambleDetLen24  = 0x06 // FSK preabmle detector length: 24-bit
	PreambleDetLen32  = 0x07 // FSK preabmle detector length: 32-bit
	AddrCompOff       = 0x00 // FSK address filtering: off
	AddrCompNode      = 0x01 // FSK address filtering: filtering on node address
	AddrCompAll       = 0x02 // FSK address filtering: filtering on node broadcast address
	PacketKnown       = 0x00 // FSK packet type: the packet length known on both side
	PacketVariable    = 0x01 // FSK packet type: the packet length on variable size
	CRC0              = 0x01 // FSK CRC type: no CRC
	CRC1              = 0x00 // FSK CRC type: CRC computed on 1 byte
	CRC2              = 0x02 // FSK CRC type: CRC computed on 2 bytes
	CRC1Inv           = 0x04 // FSK CRC type: CRC computed on 1 byte and inverted
	CRC2Inv           = 0x06 // FSK CRC type: CRC computed on 2 bytes and inverted
	WhiteningOff      = 0x00 // FSK whitening: no encoding
	WhiteningOn       = 0x01 // FSK whitening: whitening enable

	// SetCadParams
	CadOn1Symb   = 0x00 // Number of symbols used for CAD: 1
	CadOn2Symb   = 0x01 // Number of symbols used for CAD: 2
	CadOn4Symb   = 0x02 // Number of symbols used for CAD: 4
	CadOn8Symb   = 0x03 // Number of symbols used for CAD: 8
	CadOn16Symb  = 0x04 // Number of symbols used for CAD: 16
	CadExitStdby = 0x00 // After CAD is done, always exit to STDBY_RC mode
	CadExitRx    = 0x01 // After CAD is done, exit to Rx mode if activity is detected

	// GetStatus
	StatusDataAvailable = 0x04 // Packet received and data can be retrieved
	StatusCmdTimeout    = 0x06 // SPI command timed out
	StatusCmdError      = 0x08 // Invalid SPI command
	StatusCmdFailed     = 0x0A // SPI command failed to execute
	StatusCmdTxDone     = 0x0C // Packet transmission done
	StatusModeStdbyRc   = 0x20 // Chip mode: STDBY_RC
	StatusModeStdbyXosc = 0x30 // Chip mode: STDBY_XOSC
	StatusModeFs        = 0x40 // Chip mode: FS
	StatusModeRx        = 0x50 // Chip mode: RX
	StatusModeTx        = 0x60 // Chip mode: TX

	// GetDeviceErrors - device errors
	RC64KCalibErr = 0x0001 // RC64K calibration failed
	RC13MCalibErr = 0x0002 // RC13M calibration failed
	PllCalibErr   = 0x0004 // PLL calibration failed
	AdcCalibErr   = 0x0008 // ADC calibration failed
	ImgCalibErr   = 0x0010 // Image calibration failed
	XoscStartErr  = 0x0020 // Crystal oscillator failed to start
	PllLockErr    = 0x0040 // PLL failed to lock
	PaRampErr     = 0x0100 // PA ramp failed

	// LoraSyncWord
	LoraSyncWordPublic  = 0x3444 // LoRa SyncWord for public network
	LoraSyncWordPrivate = 0x1424 // LoRa SyncWord for private network (default)

	// RxGain
	RxGainPowerSaving = 0x00 // Gain used in Rx mode: power saving gain (default)
	RxGainBoosted     = 0x01 // Gain used in Rx mode: boosted gain
	PowerSavingGain   = 0x94 // Power saving gain register value
	BoostedGain       = 0x96 // Boosted gain register value

	// TX and RX operation status
	StatusDefault      = 0
	StatusTxWait       = 1
	StatusTxTimeout    = 2
	StatusTxDone       = 3
	StatusRxWait       = 4
	StatusRxContinuous = 5
	StatusRxTimeout    = 6
	StatusRxDone       = 7
	StatusHeaderErr    = 8
	StatusCrcErr       = 9
	StatusCadWait      = 10
	StatusCadDetected  = 11
	StatusCadDone      = 12

	// SX126X SPI Commands (OpCodes)
	CmdSetSleep              = 0x84
	CmdSetStandby            = 0x80
	CmdSetFs                 = 0xC1
	CmdSetTx                 = 0x83
	CmdSetRx                 = 0x82
	CmdSetRxDutyCycle        = 0x94
	CmdSetCad                = 0xC5
	CmdSetTxContinuousWave   = 0xD1
	CmdSetTxInfinitePreamble = 0xD2
	CmdSetRegulatorMode      = 0x96
	CmdCalibrate             = 0x89
	CmdCalibrateImage        = 0x98
	CmdSetPaConfig           = 0x95
	CmdSetRxTxFallbackMode   = 0x93
	CmdWriteRegister         = 0x0D
	CmdReadRegister          = 0x1D
	CmdWriteBuffer           = 0x0E
	CmdReadBuffer            = 0x1E
	CmdGetBufferStatus       = 0x13
	CmdSetDioIrqParams       = 0x08
	CmdGetIrqStatus          = 0x12
	CmdClearIrqStatus        = 0x02
	CmdSetDio2AsRfSwitchCtrl = 0x9D
	CmdSetDio3AsTcxoCtrl     = 0x97
	CmdSetRfFrequency        = 0x86
	CmdSetPacketType         = 0x8A
	CmdGetPacketType         = 0x11
	CmdSetTxParams           = 0x8E
	CmdSetModulationParams   = 0x8B
	CmdSetPacketParams       = 0x8C
	CmdGetStatus             = 0xC0
	CmdGetDeviceErrors       = 0x17
	CmdClearDeviceErrors     = 0x07
)

type Pins struct {
	Reset string
	Busy  string
	DIO   string
	TxEn  string
	RxEn  string
	CS    string
}

type Config struct {
	Enable         bool             `default:"true"`
	Modem          uint8            `default:"1"`               // lora.LoraModem / lora.FskModem
	Frequency      physic.Frequency `default:"433000000000000"` // I.e. 433000000000000 (mHz)
	DeviceType     uint16           `default:"2"`               // lora.TxPowerSX1261 / lora.TxPowerSX1262
	Bandwidth      physic.Frequency `default:"125000000000"`    // I.e. 125000000000  (mHz)
	SF             uint8            `default:"7"`               // Spreading Factor (7-12)
	CR             uint8            `default:"5"`               // Coding Rate (5 -> 4/5, 6 -> 4/6)
	LDRO           bool             `default:"false"`           // Low Data Rate Optimize - true = on, false = off
	DC_DC          bool             `default:"true"`            // DC-DC converter
	HeaderType     uint8            `default:"0"`               // lora.HeaderExplicit / lora.HeaderImplicit
	PreambleLen    uint16           `default:"12"`              // I.e. 12
	PayloadLen     uint8            `default:"32"`              // I.e. 32
	CRCType        bool             `default:"true"`            // true = on, false = off
	InvertIQ       bool             `default:"false"`           // true = on, false = off
	SyncWord       uint16           `default:"5156"`            // 0x3444 (Public Network) / 0x1424 (Private Network)
	TXPower        int8             `default:"0"`               // SX1261: -17 - +14 (dBm) ; SX1262: -9 - +22 (dBm)
	StandbyMode    uint8            `default:"0"`               // lora.StandbyRc / lora.StandbyXosc
	FrequencyRange []uint8          `default:"-"`               // [lora.CalImg430, lora.CalImg440]
	RampTime       uint8            `default:"5"`               // lora.PaRamp10 - lora.PaRamp3400u ; 0x05 = lora.PaRamp800u
	IRQMask        uint16           `default:"1023"`            // 0x03FF = IRQ All

	Pins *Pins
}

type Device struct {
	SPI    spi.Conn
	Reset  gpio.PinOut
	Busy   gpio.PinIn
	CS     gpio.PinOut
	DIO    gpio.PinIn
	TxEn   gpio.PinOut
	RxEn   gpio.PinOut
	Config *Config
}
