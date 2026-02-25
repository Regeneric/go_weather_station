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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", desc, expectedBytes, spi.TxData)
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
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", desc, expectedBytes, spi.TxData)
	}
}
