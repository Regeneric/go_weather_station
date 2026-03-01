package sgp30

type Command uint16

const (
	CmdSoftReset                   Command = 0x0006
	CmdSoftResetMSB                Command = CmdSoftReset >> 8
	CmdSoftResetLSB                Command = CmdSoftReset & 0xFF
	CmdIaqInit                     Command = 0x2003
	CmdIaqInitMSB                  Command = CmdIaqInit >> 8
	CmdIaqInitLSB                  Command = CmdIaqInit & 0xFF
	CmdMeasureIaq                  Command = 0x2008
	CmdMeasureIaqMSB               Command = CmdMeasureIaq >> 8
	CmdMeasureIaqLSB               Command = CmdMeasureIaq & 0xFF
	CmdGetIaqBaseline              Command = 0x2015
	CmdGetIaqBaselineMSB           Command = CmdGetIaqBaseline >> 8
	CmdGetIaqBaselineLSB           Command = CmdGetIaqBaseline & 0xFF
	CmdSetIaqBaseline              Command = 0x201E
	CmdSetIaqBaselineMSB           Command = CmdSetIaqBaseline >> 8
	CmdSetIaqBaselineLSB           Command = CmdSetIaqBaseline & 0xFF
	CmdSetAbsoluteHumidity         Command = 0x2061
	CmdSetAbsoluteHumidityMSB      Command = CmdSetAbsoluteHumidity >> 8
	CmdSetAbsoluteHumidityLSB      Command = CmdSetAbsoluteHumidity & 0xFF
	CmdMeasureTest                 Command = 0x2032
	CmdMeasureTestMSB              Command = CmdMeasureTest >> 8
	CmdMeasureTestLSB              Command = CmdMeasureTest & 0xFF
	CmdGetFeatureSet               Command = 0x202F
	CmdGetFeatureSetMSB            Command = CmdGetFeatureSet >> 8
	CmdGetFeatureSetLSB            Command = CmdGetFeatureSet & 0xFF
	CmdMeasureRaw                  Command = 0x2050
	CmdMeasureRawMSB               Command = CmdMeasureRaw >> 8
	CmdMeasureRawLSB               Command = CmdMeasureRaw & 0xFF
	CmdGetTvocInceptiveBaseline    Command = 0x20B3
	CmdGetTvocInceptiveBaselineMSB Command = CmdGetTvocInceptiveBaseline >> 8
	CmdGetTvocInceptiveBaselineLSB Command = CmdGetTvocInceptiveBaseline & 0xFF
	CmdSetTvocBaseline             Command = 0x2077
	CmdSetTvocBaselineMSB          Command = CmdSetTvocBaseline >> 8
	CmdSetTvocBaselineLSB          Command = CmdSetTvocBaseline & 0xFF
	CmdGetSerialId                 Command = 0x3682
	CmdGetSerialIdMSB              Command = CmdGetSerialId >> 8
	CmdGetSerialIdLSB              Command = CmdGetSerialId & 0xFF
)

type CRC uint8

const (
	CrcMask  CRC = 0x31
	CrcBase  CRC = 0xFF
	CrcMsbit CRC = 0x80
)

type Measure uint16

const (
	MeasureTest Measure = 0xD400
	FeatureSet  Measure = 0x0022
)
