package sx126x

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

func init() {
	discardHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(discardHandler))
}

// # 13.1.12 Calibrate Function
func TestCalibrate(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		param         CalibrationParam
		expectedBytes []uint8
	}{
		{
			name:          "CalibRC64k",
			desc:          "Verifies calibration of the 64 kHz RC oscillator, used for low-power operation and wake-up timers",
			param:         CalibRC64k,
			expectedBytes: []uint8{0x89, 0x01},
		},
		{
			name:          "CalibRC13M",
			desc:          "Verifies calibration of the 13 MHz RC oscillator, which provides the internal clock for the digital block",
			param:         CalibRC13M,
			expectedBytes: []uint8{0x89, 0x02},
		},
		{
			name:          "CalibPLL",
			desc:          "Verifies calibration of the Phase-Locked Loop (PLL) system to ensure frequency synthesis accuracy",
			param:         CalibPLL,
			expectedBytes: []uint8{0x89, 0x04},
		},
		{
			name:          "CalibADCPulse",
			desc:          "Verifies calibration of the ADC pulse, essential for accurate signal conversion and measurement",
			param:         CalibADCPulse,
			expectedBytes: []uint8{0x89, 0x08},
		},
		{
			name:          "CalibADCBulkN",
			desc:          "Verifies calibration of the ADC Bulk N-side to maintain linearity in the analog-to-digital conversion",
			param:         CalibADCBulkN,
			expectedBytes: []uint8{0x89, 0x10},
		},
		{
			name:          "CalibADCBulkP",
			desc:          "Verifies calibration of the ADC Bulk P-side to maintain linearity in the analog-to-digital conversion",
			param:         CalibADCBulkP,
			expectedBytes: []uint8{0x89, 0x20},
		},
		{
			name:          "CalibAll",
			desc:          "Verifies triggering calibration for all available blocks simultaneously using the full bitmask (0x3F)",
			param:         CalibAll,
			expectedBytes: []uint8{0x89, 0x3F},
		},
		{
			name:          "CalibNone",
			desc:          "Verifies that passing an empty mask results in a command sent with 0x00, effectively triggering no calibration",
			param:         CalibNone,
			expectedBytes: []uint8{0x89, 0x00},
		},
		{
			name:          "CalibRC64k|CalibRC13M",
			desc:          "Verifies bitwise OR combination: calibrating both the 64 kHz and 13 MHz RC oscillators in a single command",
			param:         (CalibRC64k | CalibRC13M),
			expectedBytes: []uint8{0x89, 0x03},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.Calibrate(tc.param)

			if err != nil {
				t.Fatalf("FAIL: %s\nCalibrate returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.13 CalibrateImage
func TestCalibrateImage(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		freq1         CalibrationImageFreq
		freq2         CalibrationImageFreq
		expectedBytes []uint8
	}{
		{
			name:          "CalImg430-CalImg440",
			desc:          "Verifies image calibration for the 430-440 MHz band, typically used in ISM regional applications",
			freq1:         CalImg430,
			freq2:         CalImg440,
			expectedBytes: []uint8{0x98, 0x6B, 0x6F},
		},
		{
			name:          "CalImg470-CalImg510",
			desc:          "Verifies image calibration for the 470-510 MHz band, commonly used for smart metering in the Chinese market",
			freq1:         CalImg470,
			freq2:         CalImg510,
			expectedBytes: []uint8{0x98, 0x75, 0x81},
		},
		{
			name:          "CalImg779-CalImg787",
			desc:          "Verifies image calibration for the 779-787 MHz band, a specific frequency range for various European sub-GHz applications",
			freq1:         CalImg779,
			freq2:         CalImg787,
			expectedBytes: []uint8{0x98, 0xC1, 0xC5},
		},
		{
			name:          "CalImg863-CalImg870",
			desc:          "Verifies image calibration for the 863-870 MHz band, the standard European ISM band (EU868)",
			freq1:         CalImg863,
			freq2:         CalImg870,
			expectedBytes: []uint8{0x98, 0xD7, 0xDB},
		},
		{
			name:          "CalImg902-CalImg928",
			desc:          "Verifies image calibration for the 902-928 MHz band, the standard North American ISM band (US915)",
			freq1:         CalImg902,
			freq2:         CalImg928,
			expectedBytes: []uint8{0x98, 0xE1, 0xE9},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.CalibrateImage(tc.freq1, tc.freq2)

			if err != nil {
				t.Fatalf("FAIL: %s\nCalibrateImage returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}
