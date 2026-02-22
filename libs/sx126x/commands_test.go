package sx126x

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type MockSPI struct {
	TxData      []uint8
	RxData      []uint8
	ReturnError error
}

func (m *MockSPI) Tx(w, r []uint8) error {
	m.TxData = append(m.TxData, w...)
	if r != nil && len(m.RxData) > 0 {
		copy(r, m.RxData)
	}
	return nil
}

func (m *MockSPI) Duplex() conn.Duplex            { return conn.Half }
func (m *MockSPI) TxPackets(p []spi.Packet) error { return nil }
func (m *MockSPI) String() string                 { return "MockSPI" }
func (m *MockSPI) Baud() physic.Frequency         { return 0 }

func init() {
	discardHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(discardHandler))
}

// # 13.1.1 SetSleep
func TestSetSleep(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		mode          SleepConfig
		expectedBytes []uint8
	}{
		{
			name:          "ColdStart",
			desc:          "Verifies setting Sleep mode with Cold Start (configuration is lost) and RTC wake-up disabled",
			mode:          SleepColdStart,
			expectedBytes: []uint8{0x84, 0x00},
		},
		{
			name:          "WarmStart",
			desc:          "Verifies setting Sleep mode with Warm Start (configuration is retained in retention memory) and RTC wake-up disabled",
			mode:          SleepWarmStart,
			expectedBytes: []uint8{0x84, 0x04},
		},
		{
			name:          "ColdStartRtc",
			desc:          "Verifies setting Sleep mode with Cold Start (configuration is lost) and RTC wake-up enabled",
			mode:          SleepColdStartRtc,
			expectedBytes: []uint8{0x84, 0x01},
		},
		{
			name:          "WarmstartRtc",
			desc:          "Verifies setting Sleep mode with Warm Start (configuration is retained in memory) and RTC wake-up enabled",
			mode:          SleepWarmStartRtc,
			expectedBytes: []uint8{0x84, 0x05},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetSleep(tc.mode)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetSleep returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.2 SetStandby
func TestSetStandby(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		mode          StandbyMode
		expectedBytes []uint8
	}{
		{
			name:          "StandbyRc",
			desc:          "Verifies setting Standby mode using the internal RC oscillator (STDBY_RC) for lower power consumption and faster wake-up",
			mode:          StandbyRc,
			expectedBytes: []uint8{0x80, 0x00},
		},
		{
			name:          "StandbyXosc",
			desc:          "Verifies setting Standby mode using the external crystal oscillator (STDBY_XOSC), which is required for precise RF operations",
			mode:          StandbyXosc,
			expectedBytes: []uint8{0x80, 0x01},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetStandby(tc.mode)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetStandby returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.3 SetFs
func TestSetFs(t *testing.T) {
	spi := MockSPI{}
	dev := Device{SPI: &spi}

	err := dev.SetFs()
	desc := "Verifies setting the Frequency Synthesis (FS) mode, which locks the PLL to the programmed frequency without enabling the RF transmitter or receiver"
	expectedBytes := []uint8{0xC1}

	if err != nil {
		t.Fatalf("FAIL: %s\nSetFs returned: %v", desc, err)
	}

	if !bytes.Equal(spi.TxData, expectedBytes) {
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", desc, expectedBytes, spi.TxData)
	}
}

// # 13.1.4 SetTx
func TestSetTx(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		timeout       uint32
		expectedBytes []uint8
	}{
		{
			name:          "TxSingle",
			desc:          "Verifies setting TX mode with a zero timeout (TxSingle), which transmits a single packet and automatically returns to Standby mode",
			timeout:       uint32(TxSingle),
			expectedBytes: []uint8{0x83, 0x00, 0x00, 0x00},
		},
		{
			name:          "TimeoutZero",
			desc:          "Verifies boundary condition: passing explicitly 0x000000 as timeout correctly configures single-packet TX mode",
			timeout:       0x000000,
			expectedBytes: []uint8{0x83, 0x00, 0x00, 0x00},
		},
		{
			name:          "TimeoutMax24bit",
			desc:          "Verifies boundary condition: setting the maximum allowed 24-bit timeout value (0xFFFFFF)",
			timeout:       0xFFFFFF,
			expectedBytes: []uint8{0x83, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "ShiftCheck",
			desc:          "Verifies correct bitwise shifting and byte order (Big-Endian) when packing a standard 24-bit timeout value into the SPI payload",
			timeout:       0x123456,
			expectedBytes: []uint8{0x83, 0x12, 0x34, 0x56},
		},
		{
			name:          "Overflow24bit",
			desc:          "Verifies that a 32-bit integer is safely truncated by masking out the highest byte, strictly enforcing the 24-bit limit of the register",
			timeout:       0xFF123456,
			expectedBytes: []uint8{0x83, 0x12, 0x34, 0x56},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetTx(tc.timeout)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetTx returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.5 SetRx
func TestSetRx(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		timeout       uint32
		expectedBytes []uint8
	}{
		{
			name:          "RxSingle",
			desc:          "Verifies setting RX mode with a zero timeout (RxSingle), which configures the modem to receive a single packet and then return to Standby mode",
			timeout:       uint32(RxSingle),
			expectedBytes: []uint8{0x82, 0x00, 0x00, 0x00},
		},
		{
			name:          "RxContinuous",
			desc:          "Verifies setting RX mode with the maximum timeout (RxContinuous), which keeps the modem in continuous reception mode",
			timeout:       uint32(RxContinuous),
			expectedBytes: []uint8{0x82, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "TimeoutZero",
			desc:          "Verifies boundary condition: passing explicitly 0x000000 correctly configures single-packet RX mode",
			timeout:       0x000000,
			expectedBytes: []uint8{0x82, 0x00, 0x00, 0x00},
		},
		{
			name:          "TimeoutMax24bit",
			desc:          "Verifies boundary condition: setting the maximum allowed 24-bit timeout value (0xFFFFFF), which acts as continuous RX mode",
			timeout:       0xFFFFFF,
			expectedBytes: []uint8{0x82, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "ShiftCheck",
			desc:          "Verifies correct bitwise shifting and byte order (Big-Endian) when packing a standard 24-bit timeout value into the SPI payload",
			timeout:       0x123456,
			expectedBytes: []uint8{0x82, 0x12, 0x34, 0x56},
		},
		{
			name:          "Overflow24bit",
			desc:          "Verifies that a 32-bit integer is safely truncated by masking out the highest byte, strictly enforcing the 24-bit limit of the register",
			timeout:       0xFF123456,
			expectedBytes: []uint8{0x82, 0x12, 0x34, 0x56},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetRx(tc.timeout)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetRx returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.6 StopTimerOnPreamble
func TestStopTimerOnPreamble(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		input         bool
		expectedBytes []uint8
	}{
		{
			name:          "TimerStop",
			desc:          "Verifies enabling the feature that automatically stops the RX timeout timer upon detecting a LoRa preamble, ensuring the modem stays in RX mode to receive the entire packet",
			input:         true,
			expectedBytes: []uint8{0x9F, 0x01},
		},
		{
			name:          "TimerNoStop",
			desc:          "Verifies disabling the preamble timer stop feature, meaning the RX timer will continue counting down and may trigger a timeout even if a preamble has been detected",
			input:         false,
			expectedBytes: []uint8{0x9F, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.StopTimerOnPreamble(tc.input)

			if err != nil {
				t.Fatalf("FAIL: %s\nStopTimerOnPreamble returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.7 SetRxDutyCycle
func TestSetRxDutyCycle(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		rxPeriod      uint32
		sleepPeriod   uint32
		expectedBytes []uint8
	}{
		{
			name:          "RxZero,SleepZero",
			desc:          "Verifies setting both RX and Sleep periods to 0, which is the absolute minimum boundary for the duty cycle configuration",
			rxPeriod:      0x000000,
			sleepPeriod:   0x000000,
			expectedBytes: []uint8{0x94, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:          "RxMax24bit,SleepMax24bit",
			desc:          "Verifies boundary condition: setting both RX and Sleep periods to their maximum allowed 24-bit values (0xFFFFFF)",
			rxPeriod:      0xFFFFFF,
			sleepPeriod:   0xFFFFFF,
			expectedBytes: []uint8{0x94, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "RxMax24bit,SleepZero",
			desc:          "Verifies configuring maximum RX period with zero Sleep period, ensuring the parameters do not overlap or interfere in the SPI payload",
			rxPeriod:      0xFFFFFF,
			sleepPeriod:   0x000000,
			expectedBytes: []uint8{0x94, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00},
		},
		{
			name:          "RxZero,SleepMax24bit",
			desc:          "Verifies configuring zero RX period with maximum Sleep period, ensuring the parameters do not overlap or interfere in the SPI payload",
			rxPeriod:      0x000000,
			sleepPeriod:   0xFFFFFF,
			expectedBytes: []uint8{0x94, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "RxZero,SleepSiftCheck",
			desc:          "Verifies packing a standard 24-bit Sleep period with a zero RX period to check bit shifting correctness for the Sleep parameter",
			rxPeriod:      0x000000,
			sleepPeriod:   0x123456,
			expectedBytes: []uint8{0x94, 0x00, 0x00, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxShiftCheck,SleepZero",
			desc:          "Verifies packing a standard 24-bit RX period with a zero Sleep period to check bit shifting correctness for the RX parameter",
			rxPeriod:      0x123456,
			sleepPeriod:   0x000000,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x00, 0x00, 0x00},
		},
		{
			name:          "RxShiftCheck,SleepShiftCheck",
			desc:          "Verifies correct bitwise shifting and byte order (Big-Endian) for both 24-bit RX and Sleep periods simultaneously",
			rxPeriod:      0x123456,
			sleepPeriod:   0x123456,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxShiftCheck,SleepMax24bit",
			desc:          "Verifies mixed boundary inputs: standard shifted RX period and maximum 24-bit Sleep period",
			rxPeriod:      0x123456,
			sleepPeriod:   0xFFFFFF,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "RxShiftCheck,SleepOverflow24bit",
			desc:          "Verifies mixed inputs: standard shifted RX period and ensures a 32-bit Sleep period is safely truncated to 24 bits",
			rxPeriod:      0x123456,
			sleepPeriod:   0xFF123456,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxMax24bit,SleepShiftCheck",
			desc:          "Verifies mixed boundary inputs: maximum 24-bit RX period and standard shifted Sleep period",
			rxPeriod:      0xFFFFFF,
			sleepPeriod:   0x123456,
			expectedBytes: []uint8{0x94, 0xFF, 0xFF, 0xFF, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxMax24bit,SleepOverflow24bit",
			desc:          "Verifies mixed inputs: maximum 24-bit RX period and ensures a 32-bit Sleep period is safely truncated to 24 bits",
			rxPeriod:      0xFFFFFF,
			sleepPeriod:   0xFF123456,
			expectedBytes: []uint8{0x94, 0xFF, 0xFF, 0xFF, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxZero,SleepOverflow24bit",
			desc:          "Verifies that a 32-bit Sleep period is safely truncated by masking out the highest byte while the RX period remains zero",
			rxPeriod:      0x000000,
			sleepPeriod:   0xFF123456,
			expectedBytes: []uint8{0x94, 0x00, 0x00, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxOverflow24bit,SleepZero",
			desc:          "Verifies that a 32-bit RX period is safely truncated by masking out the highest byte while the Sleep period remains zero",
			rxPeriod:      0xFF123456,
			sleepPeriod:   0x000000,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x00, 0x00, 0x00},
		},
		{
			name:          "RxOverflow24bit,SleepMax24bit",
			desc:          "Verifies truncation of a 32-bit RX period while setting the Sleep period to the maximum 24-bit value",
			rxPeriod:      0xFF123456,
			sleepPeriod:   0xFFFFFF,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "RxOverflow24bit,SleepShiftCheck",
			desc:          "Verifies truncation of a 32-bit RX period while using a standard shifted 24-bit Sleep period",
			rxPeriod:      0xFF123456,
			sleepPeriod:   0x123456,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxOverflow24bit,SleepOverflow24bit",
			desc:          "Verifies the ultimate safety check: both 32-bit RX and Sleep integers are safely truncated, strictly enforcing the 24-bit limit for the entire SPI frame",
			rxPeriod:      0xFF123456,
			sleepPeriod:   0xFF123456,
			expectedBytes: []uint8{0x94, 0x12, 0x34, 0x56, 0x12, 0x34, 0x56},
		},
		{
			name:          "RxOne,SleepOne",
			desc:          "Verifies the LSB (Least Significant Bit) mechanics: accurately packing the lowest non-zero value (0x000001) for both 24-bit periods",
			rxPeriod:      0x000001,
			sleepPeriod:   0x000001,
			expectedBytes: []uint8{0x94, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01},
		},
		{
			name:          "RxMsb,SleepMsb",
			desc:          "Verifies the MSB (Most Significant Bit) mechanics: accurately packing the highest single bit (0x800000) for both 24-bit periods",
			rxPeriod:      0x800000,
			sleepPeriod:   0x800000,
			expectedBytes: []uint8{0x94, 0x80, 0x00, 0x00, 0x80, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetRxDutyCycle(tc.rxPeriod, tc.sleepPeriod)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetRxDutyCycle returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.8 SetCAD
func TestSetCAD(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		modem         string
		expectedBytes []uint8
		expectError   bool
	}{
		{
			name:          "LoraModem",
			desc:          "Verifies setting the Channel Activity Detection (CAD) mode when the modem is correctly configured for LoRa, ensuring the proper SPI command is sent",
			modem:         "lora",
			expectedBytes: []uint8{0xC5},
			expectError:   false,
		},
		{
			name:          "FskModem",
			desc:          "Verifies that attempting to set Channel Activity Detection (CAD) mode while configured for FSK returns an error, as CAD is a LoRa-specific operation",
			modem:         "fsk",
			expectedBytes: nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi, Config: &Config{Modem: tc.modem}}

			err := dev.SetCAD()

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetCAD returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.9 SetTxContinuousWave
func TestSetTxContinuousWave(t *testing.T) {
	spi := MockSPI{}
	dev := Device{SPI: &spi}

	err := dev.SetTxContinuousWave()
	desc := "Verifies setting the modem into TX Continuous Wave mode, which generates an unmodulated RF carrier wave used for testing and RF certification purposes"
	expectedBytes := []uint8{0xD1}

	if err != nil {
		t.Fatalf("FAIL: %s\nSetTxContinuousWave returned: %v", desc, err)
	}

	if !bytes.Equal(spi.TxData, expectedBytes) {
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", desc, expectedBytes, spi.TxData)
	}
}

// # 13.1.10 SetTxInfinitePreamble
func TestSetTxInfinitePreamble(t *testing.T) {
	spi := MockSPI{}
	dev := Device{SPI: &spi}

	err := dev.SetTxInfinitePreamble()
	desc := "Verifies setting the modem into TX Infinite Preamble mode, which continuously transmits a preamble sequence typically used for testing or waking up sleeping receivers"
	expectedBytes := []uint8{0xD2}

	if err != nil {
		t.Fatalf("FAIL: %s\nSetTxInfinitePreamble returned: %v", desc, err)
	}

	if !bytes.Equal(spi.TxData, expectedBytes) {
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", desc, expectedBytes, spi.TxData)
	}
}

// # 13.1.11 SetRegulatorMode
func TestSetRegulatorMode(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		mode          RegulatorMode
		expectedBytes []uint8
	}{
		{
			name:          "RegulatorLdo",
			desc:          "Verifies configuring the internal Low-DropOut (LDO) linear regulator, which offers simpler power management with slightly higher power consumption",
			mode:          RegulatorLdo,
			expectedBytes: []uint8{0x96, 0x00},
		},
		{
			name:          "RegulatorDcDc",
			desc:          "Verifies configuring the high-efficiency DC-DC buck converter, which significantly reduces power consumption during TX and RX operations",
			mode:          RegulatorDcDc,
			expectedBytes: []uint8{0x96, 0x01},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetRegulatorMode(tc.mode)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetRegulatorMode returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.14 SetPaConfig
func TestSetPaConfig(t *testing.T) {
	var txMinPowerOutOfBands int8 = -50
	var txMaxPowerOutOfBands int8 = 50

	dev := &Device{}

	tests := []struct {
		name          string
		desc          string
		model         string
		baseTxPower   int8
		options       []OptionsPa
		expectedBytes []uint8
		expectError   bool
	}{
		// --- SX1261: AUTO CONFIGURATION ---
		{
			name:          "MinTxPower_Auto_1261",
			desc:          "Verifies that SX1261 auto-config correctly handles the minimum supported TX power boundary",
			model:         "1261",
			baseTxPower:   TxMinPowerSX1261,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x01, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "MaxTxPower_Auto_1261",
			desc:          "Verifies that SX1261 auto-config correctly handles the maximum supported TX power boundary",
			model:         "1261",
			baseTxPower:   TxMaxPowerSX1261,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x06, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "MinTxPowerOOB_Auto_1261",
			desc:          "Verifies that SX1261 auto-config clamps 'Out Of Bounds' low power values to the safe minimum",
			model:         "1261",
			baseTxPower:   txMinPowerOutOfBands,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x01, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "MaxTxPowerOOB_Auto_1261",
			desc:          "Verifies that SX1261 auto-config clamps 'Out Of Bounds' high power values to the safe maximum",
			model:         "1261",
			baseTxPower:   txMaxPowerOutOfBands,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x06, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus15_Auto_1261",
			desc:          "Verifies auto-config lookup for +15 dBm on SX1261",
			model:         "1261",
			baseTxPower:   15,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x06, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus14_Auto_1261",
			desc:          "Verifies auto-config lookup for +14 dBm on SX1261",
			model:         "1261",
			baseTxPower:   14,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x04, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus13_Auto_1261",
			desc:          "Verifies auto-config default lookup for +13 dBm on SX1261",
			model:         "1261",
			baseTxPower:   13,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x01, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus10_Auto_1261",
			desc:          "Verifies specific auto-config lookup for +10 dBm on SX1261",
			model:         "1261",
			baseTxPower:   10,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x01, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerZero_Auto_1261",
			desc:          "Verifies auto-config default lookup for 0 dBm on SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x01, 0x00, 0x01, 0x01},
			expectError:   false,
		},

		// --- SX1262: AUTO CONFIGURATION ---
		{
			name:          "MinTxPower_Auto_1262",
			desc:          "Verifies that SX1262 auto-config correctly handles the minimum supported TX power boundary",
			model:         "1262",
			baseTxPower:   TxMinPowerSX1262,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x02, 0x02, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "MaxTxPower_Auto_1262",
			desc:          "Verifies that SX1262 auto-config correctly handles the maximum supported TX power boundary",
			model:         "1262",
			baseTxPower:   TxMaxPowerSX1262,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x04, 0x07, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "MinTxPowerOOB_Auto_1262",
			desc:          "Verifies that SX1262 auto-config clamps 'Out Of Bounds' low power values to the safe minimum",
			model:         "1262",
			baseTxPower:   txMinPowerOutOfBands,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x02, 0x02, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "MaxTxPowerOOB_Auto_1262",
			desc:          "Verifies that SX1262 auto-config clamps 'Out Of Bounds' high power values to the safe maximum",
			model:         "1262",
			baseTxPower:   txMaxPowerOutOfBands,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x04, 0x07, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus22_Auto_1262",
			desc:          "Verifies auto-config lookup for +22 dBm on SX1262",
			model:         "1262",
			baseTxPower:   22,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x04, 0x07, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus21_Auto_1262",
			desc:          "Verifies auto-config lookup for +21 dBm on SX1262",
			model:         "1262",
			baseTxPower:   21,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x03, 0x05, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus20_Auto_1262",
			desc:          "Verifies auto-config lookup for +20 dBm on SX1262",
			model:         "1262",
			baseTxPower:   20,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x03, 0x05, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus17_Auto_1262",
			desc:          "Verifies auto-config lookup for +17 dBm on SX1262",
			model:         "1262",
			baseTxPower:   17,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x02, 0x03, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerPlus14_Auto_1262",
			desc:          "Verifies auto-config lookup for +14 dBm on SX1262",
			model:         "1262",
			baseTxPower:   14,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x02, 0x02, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "TxPowerZero_Auto_1262",
			desc:          "Verifies auto-config default lookup for 0 dBm on SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       nil,
			expectedBytes: []uint8{0x95, 0x02, 0x02, 0x00, 0x01},
			expectError:   false,
		},

		// --- MANUAL OPTIONS: SX1261 ---
		{
			name:          "PaConfigAll_1261",
			desc:          "Verifies that the multi-parameter PaConfig option correctly overrides all PA settings for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaConfig(0, 0x12, 0x34, 0x78, 0x56)},
			expectedBytes: []uint8{0x95, 0x12, 0x34, 0x56, 0x78},
			expectError:   false,
		},
		{
			name:          "PaRxPower_1261",
			desc:          "Verifies manual TX power setting via PaTxPower for SX1261 updates config without auto-tuning",
			model:         "1261",
			baseTxPower:   0x0F,
			options:       []OptionsPa{dev.PaTxPower(0x0F)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "PaDutyCycle_1261",
			desc:          "Verifies manual PaDutyCycle override for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDutyCycle(0x12)},
			expectedBytes: []uint8{0x95, 0x12, 0x00, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "PaHpMax_1261",
			desc:          "Verifies manual PaHpMax override for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaHpMax(0x34)},
			expectedBytes: []uint8{0x95, 0x00, 0x34, 0x01, 0x01},
			expectError:   false,
		},
		{
			name:          "PaDeviceSel_1261",
			desc:          "Verifies manual PaDeviceSel override for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDeviceSel(0x56)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x56, 0x01},
			expectError:   false,
		},
		{
			name:          "PaLut_1261",
			desc:          "Verifies manual PaLut override for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaLut(0x78)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x01, 0x78},
			expectError:   false,
		},
		{
			name:          "PaDutyCycle,PaHpMax_1261",
			desc:          "Verifies combining manual DutyCycle and HpMax options for SX1261",
			model:         "1261",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDutyCycle(0x12), dev.PaHpMax(0x34)},
			expectedBytes: []uint8{0x95, 0x12, 0x34, 0x01, 0x01},
			expectError:   false,
		},

		// --- MANUAL OPTIONS: SX1262 ---
		{
			name:          "PaConfigAll_1262",
			desc:          "Verifies that the multi-parameter PaConfig option correctly overrides all PA settings for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaConfig(0, 0x12, 0x34, 0x78, 0x56)},
			expectedBytes: []uint8{0x95, 0x12, 0x34, 0x56, 0x78},
			expectError:   false,
		},
		{
			name:          "PaRxPower_1262",
			desc:          "Verifies manual TX power setting via PaTxPower for SX1262 updates config without auto-tuning",
			model:         "1262",
			baseTxPower:   0x0F,
			options:       []OptionsPa{dev.PaTxPower(0x0F)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "PaDutyCycle_1262",
			desc:          "Verifies manual PaDutyCycle override for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDutyCycle(0x12)},
			expectedBytes: []uint8{0x95, 0x12, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "PaHpMax_1262",
			desc:          "Verifies manual PaHpMax override for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaHpMax(0x34)},
			expectedBytes: []uint8{0x95, 0x00, 0x34, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "PaDeviceSel_1262",
			desc:          "Verifies manual PaDeviceSel override for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDeviceSel(0x56)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x56, 0x01},
			expectError:   false,
		},
		{
			name:          "PaLut_1262",
			desc:          "Verifies manual PaLut override for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaLut(0x78)},
			expectedBytes: []uint8{0x95, 0x00, 0x00, 0x00, 0x78},
			expectError:   false,
		},
		{
			name:          "PaDutyCycle,PaHpMax_1262",
			desc:          "Verifies combining manual DutyCycle and HpMax options for SX1262",
			model:         "1262",
			baseTxPower:   0,
			options:       []OptionsPa{dev.PaDutyCycle(0x12), dev.PaHpMax(0x34)},
			expectedBytes: []uint8{0x95, 0x12, 0x34, 0x00, 0x01},
			expectError:   false,
		},

		// --- ERRORS ---
		{
			name:          "Error_UnknownModem",
			desc:          "Verifies that an unsupported modem model returns an error",
			model:         "9999",
			baseTxPower:   0,
			options:       nil,
			expectedBytes: nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			cfg := Config{Type: tc.model, TransmitPower: tc.baseTxPower}
			dev := Device{SPI: &spi, Config: &cfg}

			err := dev.SetPaConfig(tc.options...) // Options array may be empty, that's fine

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for model %q, but got nil", tc.desc, tc.model)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetPaConfig returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.1.15 SetRxTxFallbackMode
func TestSetRxTxFallbackMode(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		mode          FallbackMode
		expectedBytes []uint8
	}{
		{
			name:          "FallbackFs",
			desc:          "Verifies that the modem falls back to Frequency Synthesis (FS) mode after a packet transmission or reception, keeping the PLL locked for faster subsequent operations",
			mode:          FallbackFs,
			expectedBytes: []uint8{0x93, 0x40},
		},
		{
			name:          "FallbackStdbyXosc",
			desc:          "Verifies that the modem falls back to Standby XOSC mode after a packet transmission or reception, using the crystal oscillator for better timing accuracy than the RC oscillator",
			mode:          FallbackStdbyXosc,
			expectedBytes: []uint8{0x93, 0x30},
		},
		{
			name:          "FallbackStdbyRc",
			desc:          "Verifies that the modem falls back to Standby RC mode after a packet transmission or reception, ensuring the lowest power consumption while remaining in a ready state",
			mode:          FallbackStdbyRc,
			expectedBytes: []uint8{0x93, 0x20},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetRxTxFallbackMode(tc.mode)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetRxTxFallbackMode returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.3.1 SetDioIrqParams
func TestSetDioIrqParams(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		modem         string
		irqMask       IrqMask
		irqMasks      []IrqMask
		expectedBytes []uint8
		expectError   bool
	}{
		// LoRa IRQs - single mask (IRQ + DIO1)
		{
			name:          "TxDone_Single_LoRa",
			desc:          "Verifies that TxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a completed transmission.",
			modem:         "lora",
			irqMask:       IrqTxDone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "RxDone_Single_LoRa",
			desc:          "Verifies that RxDone is globally enabled and mapped to the DIO1 pin, ensuring the host is notified when a new packet is ready in the buffer.",
			modem:         "lora",
			irqMask:       IrqRxDone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "PreambleDetected_Single_LoRa",
			desc:          "Verifies that PreambleDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the detection of a valid preamble.",
			modem:         "lora",
			irqMask:       IrqPreambleDetected,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "SyncWordValid_Single_LoRa",
			desc:          "Verifies that attempting to enable SyncWordValid while in LoRa mode correctly returns an error, as it is an FSK-specific interrupt.",
			modem:         "lora",
			irqMask:       IrqSyncWordValid,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "HeaderValid_Single_LoRa",
			desc:          "Verifies that HeaderValid is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the reception of a valid header.",
			modem:         "lora",
			irqMask:       IrqHeaderValid,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x10, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "HeaderErr_Single_LoRa",
			desc:          "Verifies that HeaderErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a corrupted packet header.",
			modem:         "lora",
			irqMask:       IrqHeaderErr,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "CrcErr_Single_LoRa",
			desc:          "Verifies that CrcErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a payload CRC mismatch.",
			modem:         "lora",
			irqMask:       IrqCrcErr,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x40, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "CadDone_Single_LoRa",
			desc:          "Verifies that CadDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the completion of a CAD cycle.",
			modem:         "lora",
			irqMask:       IrqCadDone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x80, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "CadDetected_Single_LoRa",
			desc:          "Verifies that CadDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal activity detection in the channel.",
			modem:         "lora",
			irqMask:       IrqCadDetected,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Timeout_Single_LoRa",
			desc:          "Verifies that Timeout is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a programmed operation timeout.",
			modem:         "lora",
			irqMask:       IrqTimeout,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "All_Single_LoRa",
			desc:          "Verifies that enabling all interrupt bits (including reserved ones) is correctly rejected by the driver safety logic.",
			modem:         "lora",
			irqMask:       IrqAll,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "None_Single_LoRa",
			desc:          "Verifies that providing an empty mask correctly disables all global interrupts and clears all DIO routing.",
			modem:         "lora",
			irqMask:       IrqNone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Standard_Sane_Single_LoRa",
			desc:          "Verifies that a standard set of LoRa interrupts (TX, RX, Timeout, and Errors) is globally enabled and correctly routed to the DIO1 pin.",
			modem:         "lora",
			irqMask:       IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x02, 0x63, 0x02, 0x63, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},

		// FSK IRQs - single mask (IRQ + DIO1)
		{
			name:          "TxDone_Single_FSK",
			desc:          "Verifies that TxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:         "fsk",
			irqMask:       IrqTxDone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "RxDone_Single_FSK",
			desc:          "Verifies that RxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:         "fsk",
			irqMask:       IrqRxDone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "PreambleDetected_Single_FSK",
			desc:          "Verifies that PreambleDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:         "fsk",
			irqMask:       IrqPreambleDetected,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "SyncWordValid_Single_FSK",
			desc:          "Verifies that SyncWordValid is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a valid FSK sync word detection.",
			modem:         "fsk",
			irqMask:       IrqSyncWordValid,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "HeaderValid_Single_FSK",
			desc:          "Verifies that attempting to enable HeaderValid while in FSK mode correctly returns an error, as it is a LoRa-specific interrupt.",
			modem:         "fsk",
			irqMask:       IrqHeaderValid,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "HeaderErr_Single_FSK",
			desc:          "Verifies that attempting to enable HeaderErr while in FSK mode correctly returns an error, as it is a LoRa-specific interrupt.",
			modem:         "fsk",
			irqMask:       IrqHeaderErr,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "CrcErr_Single_FSK",
			desc:          "Verifies that CrcErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:         "fsk",
			irqMask:       IrqCrcErr,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x40, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "CadDone_Single_FSK",
			desc:          "Verifies that attempting to use CAD interrupts while in FSK mode correctly returns an error.",
			modem:         "fsk",
			irqMask:       IrqCadDone,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "CadDetected_Single_FSK",
			desc:          "Verifies that attempting to use CAD interrupts while in FSK mode correctly returns an error.",
			modem:         "fsk",
			irqMask:       IrqCadDetected,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "Timeout_Single_FSK",
			desc:          "Verifies that Timeout is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:         "fsk",
			irqMask:       IrqTimeout,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "All_Single_FSK",
			desc:          "Verifies that enabling all interrupt bits (including reserved ones) is correctly rejected by the driver safety logic for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqAll,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "None_Single_FSK",
			desc:          "Verifies that a zero mask correctly disables all global interrupts and DIO routing for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqNone,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Standard_Sane_Single_FSK",
			desc:          "Verifies that a standard set of FSK interrupts (TX, RX, Timeout, and CRC) is globally enabled and correctly routed to the DIO1 pin.",
			modem:         "fsk",
			irqMask:       IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			irqMasks:      nil,
			expectedBytes: []uint8{0x08, 0x02, 0x43, 0x02, 0x43, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},

		// LoRa IRQs - mixed
		{
			name:          "Mask_DIO1_LoRa",
			desc:          "Verifies that a custom IRQ mask is globally enabled and correctly routed exclusively to the DIO1 pin.",
			modem:         "lora",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678},
			expectedBytes: []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Mask_DIO1_DIO2_LoRa",
			desc:          "Verifies that a global IRQ mask is globally enabled and correctly routed across both DIO1 and DIO2 pins simultaneously.",
			modem:         "lora",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678, 0x4321},
			expectedBytes: []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x43, 0x21, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Mask_DIO1_DIO2_DIO3_LoRa",
			desc:          "Verifies that a global IRQ mask is globally enabled and correctly routed across all three pins (DIO1, DIO2, and DIO3) simultaneously.",
			modem:         "lora",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678, 0x4321, 0x8765},
			expectedBytes: []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x43, 0x21, 0x87, 0x65},
			expectError:   false,
		},
		{
			name:          "DIO3_Only_LoRa",
			desc:          "Verifies that an interrupt can be routed exclusively to the DIO3 pin while DIO1 and DIO2 remain disabled, ensuring correct positional packing in the SPI payload.",
			modem:         "lora",
			irqMask:       IrqTxDone,
			irqMasks:      []IrqMask{0x0000, 0x0000, IrqTxDone},
			expectedBytes: []uint8{0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "Masks_Overflow_LoRa",
			desc:          "Verifies that providing more than three DIO masks returns an error, preventing an invalid SPI transaction length.",
			modem:         "lora",
			irqMask:       IrqTxDone,
			irqMasks:      []IrqMask{IrqTxDone, IrqTxDone, IrqTxDone, IrqRxDone},
			expectedBytes: nil,
			expectError:   true,
		},

		// FSK IRQs - mixed
		{
			name:          "Mask_DIO1_FSK",
			desc:          "Verifies that a custom IRQ mask is globally enabled and correctly routed exclusively to the DIO1 pin for FSK mode.",
			modem:         "fsk",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678},
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "Mask_DIO1_DIO2_FSK",
			desc:          "Verifies that a global IRQ mask is globally enabled and correctly routed across both DIO1 and DIO2 pins simultaneously for FSK mode.",
			modem:         "fsk",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678, 0x4321},
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "Mask_DIO1_DIO2_DIO3_FSK",
			desc:          "Verifies that a global IRQ mask is globally enabled and correctly routed across all three pins (DIO1, DIO2, and DIO3) simultaneously for FSK mode.",
			modem:         "fsk",
			irqMask:       0x1234,
			irqMasks:      []IrqMask{0x5678, 0x4321, 0x8765},
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "DIO3_Only_FSK",
			desc:          "Verifies that an interrupt can be routed exclusively to the DIO3 pin while DIO1 and DIO2 remain disabled, ensuring correct positional packing in the SPI payload for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqTxDone,
			irqMasks:      []IrqMask{0x0000, 0x0000, IrqTxDone},
			expectedBytes: []uint8{0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "Masks_Overflow_FSK",
			desc:          "Verifies that providing more than three DIO masks returns an error, preventing an invalid SPI transaction length for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqTxDone,
			irqMasks:      []IrqMask{IrqTxDone, IrqTxDone, IrqTxDone, IrqRxDone},
			expectedBytes: nil,
			expectError:   true,
		},

		// Misc
		{
			name:          "Error_UnknownModem",
			desc:          "Verifies that providing an invalid modem type string returns an error, as IRQ validation logic cannot be reliably applied.",
			modem:         "generic",
			irqMask:       IrqTxDone,
			irqMasks:      nil,
			expectedBytes: nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi, Config: &Config{Modem: tc.modem}}

			err := dev.SetDioIrqParams(tc.irqMask, tc.irqMasks...)

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetDioIrqParams returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.3.3 GetIrqStatus
func TestGetIrqStatus(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		modem       string
		irqMask     IrqMask
		commands    []uint8
		tx          []uint8
		rx          []uint8
		expectError bool
	}{
		// LoRa modem
		{
			name:        "TxDone_LoRa",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the TxDone flag (bit 0) from the 16-bit interrupt status in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqTxDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_LoRa",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the RxDone flag (bit 1) from the 16-bit interrupt status in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqRxDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreambleDetected_LoRa",
			desc:        "Verifies that the driver correctly decodes the PreambleDetected flag (bit 2) from the MISO response stream while handling the byte offset correctly.",
			modem:       "lora",
			irqMask:     IrqPreambleDetected,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordValid_LoRa",
			desc:        "Verifies that the SyncWordValid flag (bit 3) is correctly decoded from the raw 16-bit IRQ status for the LoRa modem.",
			modem:       "lora",
			irqMask:     IrqSyncWordValid,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "HeaderValid_LoRa",
			desc:        "Verifies that the driver correctly decodes the HeaderValid flag (bit 4) from the 16-bit interrupt status, confirming a valid LoRa header reception.",
			modem:       "lora",
			irqMask:     IrqHeaderValid,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x10},
			expectError: false,
		},
		{
			name:        "HeaderErr_LoRa",
			desc:        "Verifies that the driver correctly decodes the HeaderErr flag (bit 5) to signal a corrupted LoRa header in the IRQ status.",
			modem:       "lora",
			irqMask:     IrqHeaderErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x20},
			expectError: false,
		},
		{
			name:        "CrcErr_LoRa",
			desc:        "Verifies that the driver correctly decodes the CrcErr flag (bit 6) from the interrupt status to signal a payload integrity failure.",
			modem:       "lora",
			irqMask:     IrqCrcErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_LoRa",
			desc:        "Verifies that the driver correctly decodes the CadDone flag (bit 7) marking the completion of a Channel Activity Detection operation.",
			modem:       "lora",
			irqMask:     IrqCadDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x80},
			expectError: false,
		},
		{
			name:        "CadDetected_LoRa",
			desc:        "Verifies that the driver correctly decodes the CadDetected flag (bit 8), which resides in the upper byte of the 16-bit IRQ status response.",
			modem:       "lora",
			irqMask:     IrqCadDetected,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x01, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_LoRa",
			desc:        "Verifies that the driver correctly decodes the Timeout flag (bit 9) from the upper byte of the 16-bit IRQ status response.",
			modem:       "lora",
			irqMask:     IrqTimeout,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_LoRa",
			desc:        "Verifies that the driver correctly handles a response with no active interrupt flags, returning a clean zero status.",
			modem:       "lora",
			irqMask:     IrqNone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_LoRa",
			desc:        "Verifies that the driver can simultaneously decode multiple active interrupt flags (TX, RX, Timeout, CRC, Header) from a single 16-bit status response.",
			modem:       "lora",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x63},
			expectError: false,
		},

		// FSK modem
		{
			name:        "TxDone_FSK",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the TxDone flag (bit 0) from the 16-bit response specifically for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTxDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_FSK",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the RxDone flag (bit 1) from the 16-bit response specifically for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqRxDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreambleDetected_FSK",
			desc:        "Verifies that the driver correctly decodes the PreambleDetected flag (bit 2) from the MISO response stream in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqPreambleDetected,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordValid_FSK",
			desc:        "Verifies that the driver correctly decodes the SyncWordValid flag (bit 3), which is a primary interrupt for FSK synchronization.",
			modem:       "fsk",
			irqMask:     IrqSyncWordValid,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "HeaderValid_FSK",
			desc:        "Verifies the decoding of bit 4 in FSK mode; ensures the driver handles the raw bitfield consistently across modem types.",
			modem:       "fsk",
			irqMask:     IrqHeaderValid,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x10},
			expectError: false,
		},
		{
			name:        "HeaderErr_FSK",
			desc:        "Verifies the decoding of bit 5 in FSK mode; ensures the driver handles the raw bitfield consistently across modem types.",
			modem:       "fsk",
			irqMask:     IrqHeaderErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x20},
			expectError: false,
		},
		{
			name:        "CrcErr_FSK",
			desc:        "Verifies that the driver correctly decodes the CrcErr flag (bit 6) to signal payload corruption in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqCrcErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_FSK",
			desc:        "Verifies the decoding of bit 7 in FSK mode; ensures consistent 16-bit status extraction.",
			modem:       "fsk",
			irqMask:     IrqCadDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x80},
			expectError: false,
		},
		{
			name:        "CadDetected_FSK",
			desc:        "Verifies the decoding of bit 8 (upper byte) in FSK mode; ensures consistent 16-bit status extraction.",
			modem:       "fsk",
			irqMask:     IrqCadDetected,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x01, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_FSK",
			desc:        "Verifies that the driver correctly decodes the Timeout flag (bit 9) from the upper byte of the 16-bit IRQ status in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTimeout,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_FSK",
			desc:        "Verifies that the driver returns an empty status when the FSK modem reports no active interrupts.",
			modem:       "fsk",
			irqMask:     IrqNone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_FSK",
			desc:        "Verifies the simultaneous decoding of a typical set of active FSK interrupt flags (TX, RX, Timeout, CRC) from the 16-bit response.",
			modem:       "fsk",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x43},
			expectError: false,
		},

		// Misc
		{
			name:        "Error_UnknownModem",
			desc:        "Verifies that providing an invalid modem type string returns an error, as IRQ validation logic cannot be reliably applied.",
			modem:       "generic",
			irqMask:     IrqTxDone,
			commands:    []uint8{0x12, OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			spi.RxData = tc.rx
			dev := Device{SPI: &spi, Config: &Config{Modem: tc.modem}}

			status, err := dev.GetIrqStatus()

			var mask uint16 = 0x03FF // Discard RFU and status bytes
			sxStatus := uint16(status) & mask
			exStatus := uint16(tc.irqMask) & mask

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nGetIrqStatus returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.tx) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.tx, spi.TxData)
			}

			if tc.modem == "lora" {
				fskBits := IrqSyncWordValid
				if sxStatus&uint16(fskBits) != 0 {
					return // It's not possible for this bit to be set in LoRa mode
				}
			}

			if tc.modem == "fsk" {
				loraBits := IrqHeaderValid | IrqHeaderErr | IrqCadDone | IrqCadDetected
				if sxStatus&uint16(loraBits) != 0 {
					return // It's not possible for this bit to be set in FSK mode
				}
			}

			if !bytes.Equal(spi.RxData, tc.rx) {
				t.Errorf("FAIL: %s\nWrong bytes read from SPI!\nExpected: [%# x]\nGot: [%# x]", tc.desc, tc.rx, spi.RxData)
			}

			if sxStatus != exStatus {
				t.Errorf("FAIL: %s\nWrong status returned from the SX126x modem!\nExpected: [% x]\nGot: [% x]", tc.desc, exStatus, sxStatus)
			}
		})
	}
}

// # 13.3.4 ClearIrqStatus
func TestClearIrqStatus(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		modem         string
		irqMask       IrqMask
		expectedBytes []uint8
		expectError   bool
	}{
		// LoRa
		{
			name:          "TxDone_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the TxDone interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqTxDone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "RxDone_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the RxDone interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqRxDone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x02},
			expectError:   false,
		},
		{
			name:          "PreableDetected_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the PreambleDetected interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqPreambleDetected,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x04},
			expectError:   false,
		},
		{
			name:          "SyncWordVAlid_LoRa",
			desc:          "Verifies that the driver returns an error when attempting to clear the FSK-specific SyncWordValid flag while configured for LoRa mode.",
			modem:         "lora",
			irqMask:       IrqSyncWordValid,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "HeaderValid_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the HeaderValid interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqHeaderValid,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x10},
			expectError:   false,
		},
		{
			name:          "HeaderErr_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the HeaderErr interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqHeaderErr,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x20},
			expectError:   false,
		},
		{
			name:          "CrcErr_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the CrcErr interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqCrcErr,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x40},
			expectError:   false,
		},
		{
			name:          "CadDone_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the CadDone interrupt flag in LoRa mode.",
			modem:         "lora",
			irqMask:       IrqCadDone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x80},
			expectError:   false,
		},
		{
			name:          "CadDetected_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the CadDetected interrupt flag, which resides in the upper byte of the payload.",
			modem:         "lora",
			irqMask:       IrqCadDetected,
			expectedBytes: []uint8{0x02, 0x00, 0x01, 0x00},
			expectError:   false,
		},
		{
			name:          "Timeout_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the Timeout interrupt flag, handling the upper byte of the payload correctly.",
			modem:         "lora",
			irqMask:       IrqTimeout,
			expectedBytes: []uint8{0x02, 0x00, 0x02, 0x00},
			expectError:   false,
		},
		{
			name:          "None_LoRa",
			desc:          "Verifies that sending an empty mask correctly results in a zeroed payload, effectively clearing no interrupts.",
			modem:         "lora",
			irqMask:       IrqNone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Standard_Sane_LoRa",
			desc:          "Verifies that the driver correctly formats the SPI command to clear a standard combination of typical LoRa interrupts simultaneously.",
			modem:         "lora",
			irqMask:       IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			expectedBytes: []uint8{0x02, 0x00, 0x02, 0x63},
			expectError:   false,
		},
		{
			name:          "All_Single_LoRa",
			desc:          "Verifies that the driver's safety logic correctly rejects an attempt to clear all bits simultaneously, as it includes restricted or invalid flags.",
			modem:         "lora",
			irqMask:       IrqAll,
			expectedBytes: nil,
			expectError:   true,
		},

		// FSK
		{
			name:          "TxDone_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the TxDone interrupt flag in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqTxDone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x01},
			expectError:   false,
		},
		{
			name:          "RxDone_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the RxDone interrupt flag in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqRxDone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x02},
			expectError:   false,
		},
		{
			name:          "PreableDetected_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the PreambleDetected interrupt flag in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqPreambleDetected,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x04},
			expectError:   false,
		},
		{
			name:          "SyncWordVAlid_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the FSK-specific SyncWordValid interrupt flag.",
			modem:         "fsk",
			irqMask:       IrqSyncWordValid,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x08},
			expectError:   false,
		},
		{
			name:          "HeaderValid_FSK",
			desc:          "Verifies that the driver returns an error when attempting to clear the LoRa-specific HeaderValid flag while configured for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqHeaderValid,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "HeaderErr_FSK",
			desc:          "Verifies that the driver returns an error when attempting to clear the LoRa-specific HeaderErr flag while configured for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqHeaderErr,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "CrcErr_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the CrcErr interrupt flag in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqCrcErr,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x40},
			expectError:   false,
		},
		{
			name:          "CadDone_FSK",
			desc:          "Verifies that the driver returns an error when attempting to clear the LoRa-specific CadDone flag while configured for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqCadDone,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "CadDetected_FSK",
			desc:          "Verifies that the driver returns an error when attempting to clear the LoRa-specific CadDetected flag while configured for FSK mode.",
			modem:         "fsk",
			irqMask:       IrqCadDetected,
			expectedBytes: nil,
			expectError:   true,
		},
		{
			name:          "Timeout_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear the Timeout interrupt flag in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqTimeout,
			expectedBytes: []uint8{0x02, 0x00, 0x02, 0x00},
			expectError:   false,
		},
		{
			name:          "None_FSK",
			desc:          "Verifies that sending an empty mask correctly results in a zeroed payload, effectively clearing no interrupts in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqNone,
			expectedBytes: []uint8{0x02, 0x00, 0x00, 0x00},
			expectError:   false,
		},
		{
			name:          "Standard_Sane_FSK",
			desc:          "Verifies that the driver correctly formats the SPI command to clear a standard combination of typical FSK interrupts simultaneously.",
			modem:         "fsk",
			irqMask:       IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			expectedBytes: []uint8{0x02, 0x00, 0x02, 0x43},
			expectError:   false,
		},
		{
			name:          "All_Single_FSK",
			desc:          "Verifies that the driver's safety logic correctly rejects an attempt to clear all bits simultaneously in FSK mode.",
			modem:         "fsk",
			irqMask:       IrqAll,
			expectedBytes: nil,
			expectError:   true,
		},

		// Misc
		{
			name:          "Error_UnknownModem",
			desc:          "Verifies that providing an invalid modem type string returns an error, as IRQ validation logic cannot be reliably applied.",
			modem:         "generic",
			irqMask:       IrqTxDone,
			expectedBytes: nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi, Config: &Config{Modem: tc.modem}}

			err := dev.ClearIrqStatus(tc.irqMask)

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nClearIrqStatus returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.3.5 SetDIO2AsRfSwitchCtrl
func TestDIO2AsRfSwitchCtrl(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		enable        bool
		expectedBytes []uint8
	}{
		{
			name:          "EnableDIO2AsRfSwitch",
			desc:          "Verifies that the driver correctly configures the DIO2 pin to automatically control the external RF switch during transmission and reception cycles.",
			enable:        true,
			expectedBytes: []uint8{0x9D, 0x01},
		},
		{
			name:          "DisableDIO2AsRfSwitch",
			desc:          "Verifies that the driver correctly disables the automated external RF switch control on the DIO2 pin.",
			enable:        false,
			expectedBytes: []uint8{0x9D, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetDIO2AsRfSwitchCtrl(tc.enable)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetDIO2AsRfSwitchCtrl returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.3.6 SetDIO3AsTCXOCtrl
func TestSetDIO3AsTCXOCtrl(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		voltage       TcxoVoltage
		timeout       uint32
		expectedBytes []uint8
	}{
		{
			name:          "DIO3Output1_6_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 1.6V with no stabilization delay.",
			voltage:       Dio3Output1_6,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output1_7_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 1.7V with no stabilization delay.",
			voltage:       Dio3Output1_7,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output1_8_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 1.8V with no stabilization delay.",
			voltage:       Dio3Output1_8,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x02, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output2_2_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 2.2V with no stabilization delay.",
			voltage:       Dio3Output2_2,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x03, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output2_4_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 2.4V with no stabilization delay.",
			voltage:       Dio3Output2_4,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x04, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output2_7_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 2.7V with no stabilization delay.",
			voltage:       Dio3Output2_7,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x05, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output3_0_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 3.0V with no stabilization delay.",
			voltage:       Dio3Output3_0,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x06, 0x00, 0x00, 0x00},
		},
		{
			name:          "DIO3Output3_3_NoTimeout",
			desc:          "Verifies that the driver correctly sets the DIO3 output voltage to 3.3V with no stabilization delay.",
			voltage:       Dio3Output3_3,
			timeout:       0x0000,
			expectedBytes: []uint8{0x97, 0x07, 0x00, 0x00, 0x00},
		},
		{
			name:          "TimeoutShift",
			desc:          "Verifies that a multi-byte timeout value is correctly split and packed into the 3-byte SPI payload in Big-Endian order.",
			voltage:       0x00,
			timeout:       0x123456,
			expectedBytes: []uint8{0x97, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name:          "TimeoutMax24bit",
			desc:          "Verifies that the maximum possible 24-bit timeout value is correctly handled and sent in the SPI transaction.",
			voltage:       0x00,
			timeout:       0x00FFFFFF,
			expectedBytes: []uint8{0x97, 0x00, 0xFF, 0xFF, 0xFF},
		},
		{
			name:          "TimeoutMinNonZero",
			desc:          "Verifies that the driver correctly processes and packs the minimum possible non-zero timeout value into the least significant byte of the payload, ensuring no data loss during bitwise operations.",
			voltage:       0x00,
			timeout:       0x00000001,
			expectedBytes: []uint8{0x97, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:          "TimeoutMSBOnly",
			desc:          "Verifies that the driver correctly shifts and packs a timeout value containing only the most significant bit of the twenty-four-bit range, ensuring the upper byte boundary is handled accurately.",
			voltage:       0x00,
			timeout:       0x00800000,
			expectedBytes: []uint8{0x97, 0x00, 0x80, 0x00, 0x00},
		},
		{
			name:          "TimeoutOverflow",
			desc:          "Verifies the driver's safety logic: any timeout value exceeding 24 bits must be capped at 0xFFFFFF to prevent corruption of the SPI frame.",
			voltage:       0x00,
			timeout:       0x12FFFFFF,
			expectedBytes: []uint8{0x97, 0x00, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetDIO3AsTCXOCtrl(tc.voltage, tc.timeout)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetDIO3AsTCXOCtrl returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent: [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}
