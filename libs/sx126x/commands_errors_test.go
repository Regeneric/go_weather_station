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

// 13.6.1 GetDeviceErrors
func TestGetDeviceErrors(t *testing.T) {
	tests := []struct {
		name       string
		desc       string
		mask       DeviceError
		tx         []uint8
		rx         []uint8
		expectedTx []uint8
	}{
		{
			name:       "RC64K_CALIB_ERR",
			desc:       "Verifies detection of the 64 kHz RC oscillator calibration failure.",
			mask:       ErrRC64KCalib,
			tx:         []uint8{uint8(CmdGetDeviceErrors), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x01},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "RC13M_CALIB_ERR",
			desc:       "Verifies detection of the 13 MHz RC oscillator calibration failure.",
			mask:       ErrRC13MCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x02},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "PLL_CALIB_ERR",
			desc:       "Verifies detection of the Phase-Locked Loop (PLL) calibration failure.",
			mask:       ErrPllCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x04},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "ADC_CALIB_ERR",
			desc:       "Verifies detection of the Analog-to-Digital Converter (ADC) calibration failure.",
			mask:       ErrAdcCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x08},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "IMG_CALIB_ERR",
			desc:       "Verifies detection of the Image Rejection calibration failure.",
			mask:       ErrImgCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x10},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "XOSC_START_ERR",
			desc:       "Verifies detection of the external crystal oscillator (XOSC) startup failure.",
			mask:       ErrXoscStart,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x20},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "PLL_LOCK_ERR",
			desc:       "Verifies detection of the PLL lock failure during frequency synthesis.",
			mask:       ErrPllLock,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x40},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "PA_RAMP_ERR",
			desc:       "Verifies detection of the Power Amplifier (PA) ramping failure during transmission.",
			mask:       ErrPaRamp,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x01, 0x00},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "XOSC_START_ERR__PLL_LOCK_ERR",
			desc:       "Verifies detection of a cascaded clock error where the external oscillator failure prevents the PLL from locking.",
			mask:       ErrXoscStart | ErrPllLock,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x60},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "RC64K__RC13M__PLL_CALIB__ADC_CALIB__IMG_CALIB",
			desc:       "Verifies detection of a total calibration failure across all internal hardware blocks.",
			mask:       ErrRC64KCalib | ErrRC13MCalib | ErrPllCalib | ErrAdcCalib | ErrImgCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x1F},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
		{
			name:       "PLL_CALIB_ERR__IMG_CALIB_ERR",
			desc:       "Verifies detection of a cascaded RF calibration error where a PLL failure directly impacts image rejection.",
			mask:       ErrPllCalib | ErrImgCalib,
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x14},
			expectedTx: []uint8{0x17, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			dev := Device{SPI: &spi}

			status, err := dev.GetDeviceErrors()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetDeviceErrors returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedTx) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedTx, spi.TxData)
			}

			if status&tc.mask == 0 {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: [%# x]\nGot:      [%# x]", tc.desc, tc.mask, status)
			}
		})
	}
}

// 13.6.2 ClearDeviceErrors
func TestClearDeviceErrors(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		clear   bool
		txBytes []uint8
	}{
		{
			name:    "InternalCacheReset",
			desc:    "Verifies sending the clear device errors command via SPI alongside a full reset of the driver's internal error cache.",
			clear:   true,
			txBytes: []uint8{0x07, 0x00, 0x00},
		},
		{
			name:    "NoInternalCacheReset",
			desc:    "Verifies sending the clear device errors command via SPI while bypassing the internal driver error cache reset.",
			clear:   false,
			txBytes: []uint8{0x07, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.ClearDeviceErrors(tc.clear)

			if err != nil {
				t.Fatalf("FAIL: %s\nClearDeviceErrors returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}
