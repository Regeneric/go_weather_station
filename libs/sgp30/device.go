package sgp30

type Register uint16

const (
	RegIaqInit                     Register = 0x2003
	RegIaqInitMSB                  Register = RegIaqInit >> 8
	RegIaqInitLSB                  Register = RegIaqInit & 0xFF
	RegMeasureIaq                  Register = 0x2008
	RegMeasureIaqMSB               Register = RegMeasureIaq >> 8
	RegMeasureIaqLSB               Register = RegMeasureIaq & 0xFF
	RegGetIaqBaseline              Register = 0x2015
	RegGetIaqBaselineMSB           Register = RegMeasureIaq >> 8
	RegGetIaqBaselineLSB           Register = RegMeasureIaq & 0xFF
	RegSetIaqBaseline              Register = 0x201E
	RegSetIaqBaselineMSB           Register = RegSetIaqBaseline >> 8
	RegSetIaqBaselineLSB           Register = RegSetIaqBaseline & 0xFF
	RegSetAbsoluteHumidity         Register = 0x2061
	RegSetAbsoluteHumidityMSB      Register = RegSetAbsoluteHumidity >> 8
	RegSetAbsoluteHumidityLSB      Register = RegSetAbsoluteHumidity & 0xFF
	RegMeasureTest                 Register = 0x2032
	RegMeasureTestMSB              Register = RegMeasureTest >> 8
	RegMeasureTestLSB              Register = RegMeasureTest & 0xFF
	RegGetFeatureSet               Register = 0x202F
	RegGetFeatureSetMSB            Register = RegGetFeatureSet >> 8
	RegGetFeatureSetLSB            Register = RegGetFeatureSet & 0xFF
	RegMeasureRaw                  Register = 0x2050
	RegMeasureRawMSB               Register = RegMeasureRaw >> 8
	RegMeasureRawLSB               Register = RegMeasureRaw & 0xFF
	RegGetTvocInceptiveBaseline    Register = 0x20B3
	RegGetTvocInceptiveBaselineMSB Register = RegGetTvocInceptiveBaseline >> 8
	RegGetTvocInceptiveBaselineLSB Register = RegGetTvocInceptiveBaseline & 0xFF
	RegSetTvocBaseline             Register = 0x2077
	RegSetTvocBaselineMSB          Register = RegSetTvocBaseline >> 8
	RegSetTvocBaselineLSB          Register = RegSetTvocBaseline & 0xFF
)
