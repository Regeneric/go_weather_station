package sx126x

//go:generate stringer -type=Register
type Register uint16

const (
	RegFskWhiteningInitialMsb Register = 0x06B8
	RegFskCrcInitialMsb       Register = 0x06BC
	RegFskSyncWord0           Register = 0x06C0
	RegFskNodeAddress         Register = 0x06CD
	RegIqPolaritySetup        Register = 0x0736
	RegLoraSyncWordMsb        Register = 0x0740
	RegTxModulation           Register = 0x0889
	RegRxGain                 Register = 0x08AC
	RegTxClampConfig          Register = 0x08D8
	RegOcpConfiguration       Register = 0x08E7
	RegRtcControl             Register = 0x0902
	RegXtaTrim                Register = 0x0911
	RegXtbTrim                Register = 0x0912
	RegEventMask              Register = 0x0944
)

//go:generate stringer -type=OpCode
type OpCode uint8

const (
	// SX126X SPI Commands (OpCodes)
	CmdSetSleep              OpCode = 0x84
	CmdSetStandby            OpCode = 0x80
	CmdSetFs                 OpCode = 0xC1
	CmdSetTx                 OpCode = 0x83
	CmdSetRx                 OpCode = 0x82
	CmdSetRxDutyCycle        OpCode = 0x94
	CmdSetCad                OpCode = 0xC5
	CmdSetTxContinuousWave   OpCode = 0xD1
	CmdSetTxInfinitePreamble OpCode = 0xD2
	CmdSetRegulatorMode      OpCode = 0x96
	CmdCalibrate             OpCode = 0x89
	CmdCalibrateImage        OpCode = 0x98
	CmdSetPaConfig           OpCode = 0x95
	CmdSetRxTxFallbackMode   OpCode = 0x93
	CmdWriteRegister         OpCode = 0x0D
	CmdReadRegister          OpCode = 0x1D
	CmdWriteBuffer           OpCode = 0x0E
	CmdReadBuffer            OpCode = 0x1E
	CmdGetBufferStatus       OpCode = 0x13
	CmdSetDioIrqParams       OpCode = 0x08
	CmdGetIrqStatus          OpCode = 0x12
	CmdClearIrqStatus        OpCode = 0x02
	CmdSetDio2AsRfSwitchCtrl OpCode = 0x9D
	CmdSetDio3AsTcxoCtrl     OpCode = 0x97
	CmdSetRfFrequency        OpCode = 0x86
	CmdSetPacketType         OpCode = 0x8A
	CmdGetPacketType         OpCode = 0x11
	CmdSetTxParams           OpCode = 0x8E
	CmdSetModulationParams   OpCode = 0x8B
	CmdSetPacketParams       OpCode = 0x8C
	CmdGetStatus             OpCode = 0xC0
	CmdGetStats              OpCode = 0x10
	CmdGetDeviceErrors       OpCode = 0x17
	CmdClearDeviceErrors     OpCode = 0x07
	CmdSetBufferBaseAddress  OpCode = 0x8F
	CmdStopOnPreamble        OpCode = 0x9F
	CmdSetCadParams          OpCode = 0x88
	CmdSetSymbNumTimeout     OpCode = 0xA0
	CmdGetPacketStatus       OpCode = 0x14
	CmdGetPacketRssi         OpCode = 0x15
	CmdResetStats            OpCode = 0x00
	CmdResetErrors           OpCode = 0x07
	OpCodeNop                       = 0x00
)

//go:generate stringer -type=SleepConfig
type SleepConfig uint8

const (
	SleepColdStart    = 0x00 // Cold start, configuration is lost (default)
	SleepWarmStart    = 0x04 // Warm start, configuration is retained
	SleepColdStartRtc = 0x01 // Cold start and wake on RTC timeout
	SleepWarmStartRtc = 0x05 // Warm start and wake on RTC timeout
)

//go:generate stringer -type=StandbyMode
type StandbyMode uint8

const (
	StandbyRc   StandbyMode = 0x00 // 13 MHz RC oscillator
	StandbyXosc StandbyMode = 0x01 // 32 MHz crystal oscillator
)

//go:generate stringer -type=RegulatorMode
type RegulatorMode uint8

const (
	RegulatorLdo  RegulatorMode = 0x00 // LDO (default)
	RegulatorDcDc RegulatorMode = 0x01 // DC-DC
)

//go:generate stringer -type=TxRxTimeout
type TxRxTimeout uint32

const (
	// timeouts = timeout * 15.625 us (24-bit).
	TxSingle     TxRxTimeout = 0x000000 // No timeout (Tx single mode)
	RxSingle     TxRxTimeout = 0x000000 // No timeout (Rx single mode)
	RxContinuous TxRxTimeout = 0xFFFFFF // Infinite (Rx continuous mode)
)

//go:generate stringer -type=FallbackMode
type FallbackMode uint8

const (
	FallbackFs        FallbackMode = 0x40 // FS mode
	FallbackStdbyXosc FallbackMode = 0x30 // Crystal oscillator
	FallbackStdbyRc   FallbackMode = 0x20 // RC oscillator (default)
)

//go:generate stringer -type=CalibrationImageFreq
type CalibrationImageFreq uint8

const (
	CalImg430 CalibrationImageFreq = 0x6B // 430 - 440 MHz
	CalImg440 CalibrationImageFreq = 0x6F
	CalImg470 CalibrationImageFreq = 0x75 // 470 - 510 MHz
	CalImg510 CalibrationImageFreq = 0x81
	CalImg779 CalibrationImageFreq = 0xC1 // 779 - 787 MHz
	CalImg787 CalibrationImageFreq = 0xC5
	CalImg863 CalibrationImageFreq = 0xD7 // 863 - 870 MHz
	CalImg870 CalibrationImageFreq = 0xDB
	CalImg902 CalibrationImageFreq = 0xE1 // 902 - 928 MHz
	CalImg928 CalibrationImageFreq = 0xE9
)

//go:generate stringer -type=CalibrationParam
type CalibrationParam uint8

const (
	CalibNone     CalibrationParam = 0x00
	CalibRC64k    CalibrationParam = 0x01
	CalibRC13M    CalibrationParam = 0x02
	CalibPLL      CalibrationParam = 0x04
	CalibADCPulse CalibrationParam = 0x08
	CalibADCBulkN CalibrationParam = 0x10
	CalibADCBulkP CalibrationParam = 0x20
	CalibAll      CalibrationParam = 0x3F // All of them (0x01 | ... | 0x20)
)

const (
	RfFrequencyXtal = 32000000 // XTAL frequency used for RF frequency calculation
	RfFrequencyNom  = 33554432 // Used for RF frequency calculation
)

//go:generate stringer -type=PaConfigDeviceSel
type PaConfigDeviceSel uint8

const (
	TxPowerSX1261 PaConfigDeviceSel = 0x01
	TxPowerSX1262 PaConfigDeviceSel = 0x00
)

const (
	// Limit in dBm
	TxMaxPowerSX1261 int8 = 15
	TxMinPowerSX1261 int8 = -17
	TxMaxPowerSX1262 int8 = 22
	TxMinPowerSX1262 int8 = -9
)

//go:generate stringer -type=RampTime
type RampTime uint8

const (
	PaRamp10u   RampTime = 0x00 // Ramp time 10 us
	PaRamp20u   RampTime = 0x01 // Ramp time 20 us
	PaRamp40u   RampTime = 0x02 // Ramp time 40 us
	PaRamp80u   RampTime = 0x03 // Ramp time 80 us
	PaRamp200u  RampTime = 0x04 // Ramp time 200 us
	PaRamp800u  RampTime = 0x05 // Ramp time 800 us
	PaRamp1700u RampTime = 0x06 // Ramp time 1700 us
	PaRamp3400u RampTime = 0x07 // Ramp time 3400 us
)

//go:generate stringer -type=IrqMask
type IrqMask uint16

const (
	IrqTxDone           IrqMask = 0x0001 // Packet transmission completed
	IrqRxDone           IrqMask = 0x0002 // Packet received
	IrqPreambleDetected IrqMask = 0x0004 // Preamble detected
	IrqSyncWordValid    IrqMask = 0x0008 // Valid sync word detected
	IrqHeaderValid      IrqMask = 0x0010 // Valid LoRa header received
	IrqHeaderErr        IrqMask = 0x0020 // LoRa header CRC error
	IrqCrcErr           IrqMask = 0x0040 // Wrong CRC received
	IrqCadDone          IrqMask = 0x0080 // Channel activity detection finished
	IrqCadDetected      IrqMask = 0x0100 // Channel activity detected
	IrqTimeout          IrqMask = 0x0200 // Rx or Tx timeout
	IrqAll              IrqMask = 0x03FF // All interupts
	IrqNone             IrqMask = 0x0000 // No interupts
)

//go:generate stringer -type=Dio2Mode
type Dio2Mode uint8

const (
	Dio2AsIrq      = 0x00 // IRQ
	Dio2AsRfSwitch = 0x01 // RF switch control
)

//go:generate stringer -type=TcxoVoltage
type TcxoVoltage uint8

const (
	Dio3Output1_6 TcxoVoltage = 0x00 // 1.6V
	Dio3Output1_7 TcxoVoltage = 0x01 // 1.7V
	Dio3Output1_8 TcxoVoltage = 0x02 // 1.8V
	Dio3Output2_2 TcxoVoltage = 0x03 // 2.2V
	Dio3Output2_4 TcxoVoltage = 0x04 // 2.4V
	Dio3Output2_7 TcxoVoltage = 0x05 // 2.7V
	Dio3Output3_0 TcxoVoltage = 0x06 // 3.0V
	Dio3Output3_3 TcxoVoltage = 0x07 // 3.3V
)

//go:generate stringer -type=TcxoDelay
type TcxoDelay uint32

const (
	TcxoDelay2_5 TcxoDelay = 0x0140 // 2.5 ms
	TcxoDelay5   TcxoDelay = 0x0280 // 5 ms
	TcxoDelay10  TcxoDelay = 0x0560 // 10 ms
)

//go:generate stringer -type=PacketType
type PacketType uint8

const (
	PacketTypeGFSK PacketType = 0x00
	PacketTypeLoRa PacketType = 0x01
)

//go:generate stringer -type=LoRaBandwidth
type LoRaBandwidth uint8

const (
	LoRaBW_7_8   LoRaBandwidth = 0x00 // 7.8 kHz
	LoRaBW_10_4  LoRaBandwidth = 0x08 // 10.4 kHz
	LoRaBW_15_6  LoRaBandwidth = 0x01 // 15.6 kHz
	LoRaBW_20_8  LoRaBandwidth = 0x09 // 20.8 kHz
	LoRaBW_31_25 LoRaBandwidth = 0x02 // 31.25 kHz
	LoRaBW_41_7  LoRaBandwidth = 0x0A // 41.7 kHz
	LoRaBW_62_5  LoRaBandwidth = 0x03 // 62.5 kHz
	LoRaBW_125   LoRaBandwidth = 0x04 // 125.0 kHz
	LoRaBW_250   LoRaBandwidth = 0x05 // 250.0 kHz
	LoRaBW_500   LoRaBandwidth = 0x06 // 500.0 kHz
)

//go:generate stringer -type=LoRaCodingRate
type LoRaCodingRate uint8

const (
	LoRaCR_4_5 LoRaCodingRate = 0x01 // 4/5
	LoRaCR_4_6 LoRaCodingRate = 0x02 // 4/6
	LoRaCR_4_7 LoRaCodingRate = 0x03 // 4/7
	LoRaCR_4_8 LoRaCodingRate = 0x04 // 4/8
)

//go:generate stringer -type=LoRaLowDataRateOptimize
type LoRaLowDataRateOptimize uint8

const (
	LDRO_OFF LoRaLowDataRateOptimize = 0x00 // LoRa low data rate optimization: disabled
	LDRO_ON  LoRaLowDataRateOptimize = 0x01 // LoRa low data rate optimization: enabled
)

//go:generate stringer -type=:LoRaHeaderType
type LoRaHeaderType uint8

const (
	HeaderExplicit LoRaHeaderType = 0x00
	HeaderImplicit LoRaHeaderType = 0x01
)

//go:generate stringer -type=LoRaCrcMode
type LoRaCrcMode uint8

const (
	CrcOff LoRaCrcMode = 0x00
	CrcOn  LoRaCrcMode = 0x01
)

//go:generate stringer -type=LoRaIQMode
type LoRaIQMode uint8

const (
	IqStandard LoRaIQMode = 0x00
	IqInverted LoRaIQMode = 0x01
)

//go:generate stringer -type=FskPulseShape
type FskPulseShape uint8

const (
	PulseNoFilter      FskPulseShape = 0x00 // No filter
	PulseGaussianBt0_3 FskPulseShape = 0x08 // Gaussian BT 0.3
	PulseGaussianBt0_5 FskPulseShape = 0x09 // Gaussian BT 0.5
	PulseGaussianBt0_7 FskPulseShape = 0x0A // Gaussian BT 0.7
	PulseGaussianBt1   FskPulseShape = 0x0B // Gaussian BT 1.0
)

//go:generate stringer -type=FskBandwidth
type FskBandwidth uint8

const (
	FskBW_4800   FskBandwidth = 0x1F // 4.8 kHz
	FskBW_5800   FskBandwidth = 0x17 // 5.8 kHz
	FskBW_7300   FskBandwidth = 0x0F // 7.3 kHz
	FskBW_9700   FskBandwidth = 0x1E // 9.7 kHz
	FskBW_11700  FskBandwidth = 0x16 // 11.7 kHz
	FskBW_14600  FskBandwidth = 0x0E // 14.6 kHz
	FskBW_19500  FskBandwidth = 0x1D // 19.5 kHz
	FskBW_23400  FskBandwidth = 0x15 // 23.4 kHz
	FskBW_29300  FskBandwidth = 0x0D // 29.3 kHz
	FskBW_39000  FskBandwidth = 0x1C // 39 kHz
	FskBW_46900  FskBandwidth = 0x14 // 46.9 kHz
	FskBW_58600  FskBandwidth = 0x0C // 58.6 kHz
	FskBW_78200  FskBandwidth = 0x1B // 78.2 kHz
	FskBW_93800  FskBandwidth = 0x13 // 93.8 kHz
	FskBW_117300 FskBandwidth = 0x0B // 117.3 kHz
	FskBW_156200 FskBandwidth = 0x1A // 156.2 kHz
	FskBW_187200 FskBandwidth = 0x12 // 187.2 kHz
	FskBW_234300 FskBandwidth = 0x0A // 234.3 kHz
	FskBW_312000 FskBandwidth = 0x19 // 312 kHz
	FskBW_373600 FskBandwidth = 0x11 // 373.6 kHz
	FskBW_467000 FskBandwidth = 0x09 // 467 kHz
)

//go:generate stringer -type=FskPreambleDetector
type FskPreambleDetector uint8

const (
	PreambleDetLenOff FskPreambleDetector = 0x00 // FSK preabmle detector length: off
	PreambleDetLen8   FskPreambleDetector = 0x04 // FSK preabmle detector length: 8-bit
	PreambleDetLen16  FskPreambleDetector = 0x05 // FSK preabmle detector length: 16-bit
	PreambleDetLen24  FskPreambleDetector = 0x06 // FSK preabmle detector length: 24-bit
	PreambleDetLen32  FskPreambleDetector = 0x07 // FSK preabmle detector length: 32-bit
)

const (
	FskBitrateMin uint64 = 600
	FskBitrateMax uint64 = 300000
)

//go:generate stringer -type=FskAddressComp
type FskAddressComp uint8

const (
	AddrCompOff  FskAddressComp = 0x00
	AddrCompNode FskAddressComp = 0x01
	AddrCompAll  FskAddressComp = 0x02
)

//go:generate stringer -type=FskPacketLengthMode
type FskPacketLengthMode uint8

const (
	PacketKnown    FskPacketLengthMode = 0x00
	PacketVariable FskPacketLengthMode = 0x01
)

//go:generate stringer -type=FskCrcType
type FskCrcType uint8

const (
	CRC0    FskCrcType = 0x01 // No CRC
	CRC1    FskCrcType = 0x00 // CRC 1 byte
	CRC2    FskCrcType = 0x02 // CRC 2 bytes
	CRC1Inv FskCrcType = 0x04 // CRC 1 byte inverted
	CRC2Inv FskCrcType = 0x06 // CRC 2 bytes inverted
)

//go:generate stringer -type=FskWhitening
type FskWhitening uint8

const (
	WhiteningOff FskWhitening = 0x00
	WhiteningOn  FskWhitening = 0x01
)

//go:generate stringer -type=CadSymbolNum
type CadSymbolNum uint8

const (
	CadOn1Symb  CadSymbolNum = 0x00 // Number of symbols used for CAD: 1
	CadOn2Symb  CadSymbolNum = 0x01 // Number of symbols used for CAD: 2
	CadOn4Symb  CadSymbolNum = 0x02 // Number of symbols used for CAD: 4
	CadOn8Symb  CadSymbolNum = 0x03 // Number of symbols used for CAD: 8
	CadOn16Symb CadSymbolNum = 0x04 // Number of symbols used for CAD: 16
)

//go:generate stringer -type=CadExitMode
type CadExitMode uint8

const (
	CadExitStdby CadExitMode = 0x00 // Always exit to STDBY_RC
	CadExitRx    CadExitMode = 0x01 // Exit to Rx if detected
)

//go:generate stringer -type=StatusMode
type StatusMode uint8

const (
	StatusModeStdbyRc   StatusMode = 0x20 // Chip mode: STDBY_RC
	StatusModeStdbyXosc StatusMode = 0x30 // Chip mode: STDBY_XOSC
	StatusModeFs        StatusMode = 0x40 // Chip mode: FS
	StatusModeRx        StatusMode = 0x50 // Chip mode: RX
	StatusModeTx        StatusMode = 0x60 // Chip mode: TX
)

//go:generate stringer -type=CommandStatus
type CommandStatus uint8

const (
	StatusDataAvailable CommandStatus = 0x04 // Packet received and data can be retrieved
	StatusCmdTimeout    CommandStatus = 0x06 // SPI command timed out
	StatusCmdError      CommandStatus = 0x08 // Invalid SPI command
	StatusCmdFailed     CommandStatus = 0x0A // SPI command failed to execute
	StatusCmdTxDone     CommandStatus = 0x0C // Packet transmission done
)

//go:generate stringer -type=DeviceError
type DeviceError uint16

const (
	ErrRC64KCalib DeviceError = 0x0001 // RC64K calibration failed
	ErrRC13MCalib DeviceError = 0x0002 // RC13M calibration failed
	ErrPllCalib   DeviceError = 0x0004 // PLL calibration failed
	ErrAdcCalib   DeviceError = 0x0008 // ADC calibration failed
	ErrImgCalib   DeviceError = 0x0010 // Image calibration failed
	ErrXoscStart  DeviceError = 0x0020 // Crystal oscillator failed to start
	ErrPllLock    DeviceError = 0x0040 // PLL failed to lock
	ErrPaRamp     DeviceError = 0x0100 // PA ramp failed
)

func (e DeviceError) Has(flag DeviceError) bool {
	return e&flag != 0
}

const (
	LoraSyncWordPublic  uint16 = 0x3444 // LoRa SyncWord for public network
	LoraSyncWordPrivate uint16 = 0x1424 // LoRa SyncWord for private network (default)
)

//go:generate stringer -type=RxGain
type RxGain uint8

const (
	RxGainPowerSaving RxGain = 0x00 // Gain used in Rx mode: power saving gain (default)
	RxGainBoosted     RxGain = 0x01 // Gain used in Rx mode: boosted gain

	RxGainRegPowerSaving uint8 = 0x94
	RxGainRegBoosted     uint8 = 0x96
)

//go:generate stringer -type=DriverStatus
type DriverStatus int

const (
	StatusDefault DriverStatus = iota
	StatusTxWait
	StatusTxTimeout
	StatusTxDone
	StatusRxWait
	StatusRxContinuous
	StatusRxTimeout
	StatusRxDone
	StatusHeaderErr
	StatusCrcErr
	StatusCadWait
	StatusCadDetected
	StatusCadDone
)
