package sgp30

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"testing"
)

func init() {
	discardHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(discardHandler))
}

func TestValidateCRC(t *testing.T) {
	tests := []struct {
		name      string
		desc      string
		data      []uint8
		eco2Error bool
		tvocError bool
	}{
		{
			name:      "CO2_0_CRC_Ok__TVOC_0_CRC_Ok",
			desc:      "Zero values for both measurements with correct checksums.",
			data:      []uint8{0x00, 0x00, 0x81, 0x00, 0x00, 0x81},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_Max16bit_CRC_Ok__TVOC_Max16bit_CRC_Ok",
			desc:      "Maximum values for both measurements with correct checksums.",
			data:      []uint8{0xFF, 0xFF, 0xAC, 0xFF, 0xFF, 0xAC},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_Shift_CRC_Ok__TVOC_Shift_CRC_Ok",
			desc:      "Different, non-zero values to verify the machine isn't mixing up eCO2 and TVOC blocks.",
			data:      []uint8{0x12, 0x34, 0x37, 0x56, 0x78, 0x7D},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_128_CRC_Ok__TVOC_256_CRC_Ok",
			desc:      "Specific measurement values with correct checksums.",
			data:      []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x75},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_128_CRC_Error__TVOC_256_CRC_Ok",
			desc:      "Bad CRC for the first block only. Expected eCO2 validation error.",
			data:      []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x75},
			eco2Error: true,
			tvocError: false,
		},
		{
			name:      "CO2_128_CRC_Ok__TVOC_256_CRC_Error",
			desc:      "Bad CRC for the second block only. Expected TVOC validation error.",
			data:      []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x00},
			eco2Error: false,
			tvocError: true,
		},
		{
			name:      "CO2_128_CRC_Error__TVOC_256_CRC_Error",
			desc:      "Both checksums are corrupted. Expected errors for both measurements.",
			data:      []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			eco2Error: true,
			tvocError: true,
		},
		{
			name:      "CO2_0_CRC_Ok",
			desc:      "Zero values measurements with correct checksums.",
			data:      []uint8{0x00, 0x00, 0x81},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_Max16bit_CRC_Ok",
			desc:      "Maximum values for measurements with correct checksums.",
			data:      []uint8{0xFF, 0xFF, 0xAC},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_Shift_CRC_Ok",
			desc:      "Different, non-zero values to verify the machine isn't mixing up bytes.",
			data:      []uint8{0x12, 0x34, 0x37},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_128_CRC_Ok",
			desc:      "Specific measurement values with correct checksums - single block.",
			data:      []uint8{0x00, 0x80, 0xFB},
			eco2Error: false,
			tvocError: false,
		},
		{
			name:      "CO2_128_CRC_Error",
			desc:      "Bad CRC, expected eCO2 validation error.",
			data:      []uint8{0x00, 0x80, 0x00},
			eco2Error: true,
			tvocError: false,
		},
		{
			name:      "DataFrameTooShort",
			desc:      "Frame cut down to 5 bytes (broken transmission). Should throw an error before TVOC validation.",
			data:      []uint8{0x00, 0x80, 0x00, 0x01, 0x00},
			eco2Error: true,
			tvocError: true,
		},
		{
			name:      "DataFrameTooLong",
			desc:      "Frame extended to 7 bytes (garbage on the I2C bus). The validation function should reject it.",
			data:      []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00, 0xFF},
			eco2Error: true,
			tvocError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eco2 := tc.data[0:3]
			err := validateCRC(eco2)

			if tc.eco2Error == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for eCO2 CRC validaton but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\neCO2 CRC returned: %v", tc.desc, err)
			}

			if len(tc.data) == 3 {
				return
			}

			tvoc := tc.data[3:6]
			err = validateCRC(tvoc)

			if tc.tvocError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for TVOC CRC validaton but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\neTVOC CRC returned: %v", tc.desc, err)
			}

		})
	}
}

// Table 10 Measurement commands.
func TestIaqInit(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "SGP30_IAQ_Init_Success",
			desc:        "Correctly send the IAQ Init command",
			txBytes:     []uint8{0x20, 0x03},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "SGP30_IAQ_Init_Failed",
			desc:        "Simulate I2C bus error during IAQ Init",
			txBytes:     nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError}
			dev := Device{I2C: &i2c}

			err := dev.IaqInit()

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nIaqInit returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes sent to I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, i2c.TxData)
			}
		})
	}
}

// Table 10 Measurement commands.
func TestMeasureIaq(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		rxBytes     []uint8
		rx          []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "CO2_0_CRC_Ok__TVOC_0_CRC_Ok",
			desc:        "Minimum values for both measurements with correct checksums.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x00, 0x81, 0x00, 0x00, 0x81},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Max16bit_CRC_Ok__TVOC_Max16bit_CRC_Ok",
			desc:        "Maximum values for both measurement blocks with correct checksums.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0xFF, 0xFF, 0xAC, 0xFF, 0xFF, 0xAC},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Shift_CRC_Ok__TVOC_Shift_CRC_Ok",
			desc:        "Different non-zero values to ensure the driver does not mix up eCO2 and TVOC data.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x12, 0x34, 0x37, 0x56, 0x78, 0x7D},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Ok",
			desc:        "Valid standard sensor readings with correct checksums.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x75},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Ok",
			desc:        "Invalid checksum in the first measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x75},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Error",
			desc:        "Invalid checksum in the second measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Error",
			desc:        "Invalid checksums for both measurements. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "DataBufferTooShort",
			desc:        "Data buffer provided by the caller is too small. The function should reject it.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 5),
			expectError: true,
		},
		{
			name:        "DataBufferTooLong",
			desc:        "Data buffer provided by the caller is larger than needed. The function should reject it.",
			txBytes:     []uint8{0x20, 0x08},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 7),
			expectError: true,
		},
		{
			name:        "I2C_HardwareError",
			desc:        "Simulate hardware bus error during measurement.",
			txBytes:     nil,
			rxBytes:     nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError, RxData: tc.rxBytes}
			dev := Device{I2C: &i2c}

			err := dev.MeasureIaq(tc.rx)

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nMeasureIaq returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to I2C!\nExpected: [%# x ]\nSent:     [%# x ]", tc.desc, tc.txBytes, i2c.TxData)
			}

			if !bytes.Equal(i2c.RxData, tc.rx) {
				t.Errorf("FAIL: %s\nWrong bytes received from I2C!\nExpected: [%# x ]\nSent:     [%# x ]", tc.desc, tc.rx, i2c.RxData)
			}
		})
	}
}

func TestGetIaqBaseline(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		rxBytes     []uint8
		rx          []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "CO2_0_CRC_Ok__TVOC_0_CRC_Ok",
			desc:        "Minimum values for both measurements with correct checksums.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x00, 0x81, 0x00, 0x00, 0x81},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Max16bit_CRC_Ok__TVOC_Max16bit_CRC_Ok",
			desc:        "Maximum values for both measurement blocks with correct checksums.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0xFF, 0xFF, 0xAC, 0xFF, 0xFF, 0xAC},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Shift_CRC_Ok__TVOC_Shift_CRC_Ok",
			desc:        "Different non-zero values to ensure the driver does not mix up eCO2 and TVOC data.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x12, 0x34, 0x37, 0x56, 0x78, 0x7D},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Ok",
			desc:        "Valid standard sensor readings with correct checksums.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x75},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Ok",
			desc:        "Invalid checksum in the first measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x75},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Error",
			desc:        "Invalid checksum in the second measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Error",
			desc:        "Invalid checksums for both measurements. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 6),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "DataBufferTooShort",
			desc:        "Data buffer provided by the caller is too small. The function should reject it.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 5),
			expectError: true,
		},
		{
			name:        "DataBufferTooLong",
			desc:        "Data buffer provided by the caller is larger than needed. The function should reject it.",
			txBytes:     []uint8{0x20, 0x15},
			rxBytes:     []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			rx:          make([]uint8, 7),
			expectError: true,
		},
		{
			name:        "I2C_HardwareError",
			desc:        "Simulate hardware bus error during measurement.",
			txBytes:     nil,
			rxBytes:     nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError, RxData: tc.rxBytes}
			dev := Device{I2C: &i2c}

			err := dev.GetIaqBaseline(tc.rx)

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nGetIaqBaseline returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, i2c.TxData)
			}

			if !bytes.Equal(i2c.RxData, tc.rx) {
				t.Errorf("FAIL: %s\nWrong bytes received from I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.rx, i2c.RxData)
			}
		})
	}
}

func TestSetIaqBaseline(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		tx          []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "CO2_0_CRC_Ok__TVOC_0_CRC_Ok",
			desc:        "Minimum values for both measurements with correct checksums.",
			txBytes:     []uint8{0x20, 0x1E, 0x00, 0x00, 0x81, 0x00, 0x00, 0x81},
			tx:          []uint8{0x00, 0x00, 0x81, 0x00, 0x00, 0x81},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Max16bit_CRC_Ok__TVOC_Max16bit_CRC_Ok",
			desc:        "Maximum values for both measurement blocks with correct checksums.",
			txBytes:     []uint8{0x20, 0x1E, 0xFF, 0xFF, 0xAC, 0xFF, 0xFF, 0xAC},
			tx:          []uint8{0xFF, 0xFF, 0xAC, 0xFF, 0xFF, 0xAC},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_Shift_CRC_Ok__TVOC_Shift_CRC_Ok",
			desc:        "Different non-zero values to ensure the driver does not mix up eCO2 and TVOC data.",
			txBytes:     []uint8{0x20, 0x1E, 0x56, 0x78, 0x7D, 0x12, 0x34, 0x37},
			tx:          []uint8{0x12, 0x34, 0x37, 0x56, 0x78, 0x7D},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Ok",
			desc:        "Valid standard sensor readings with correct checksums.",
			txBytes:     []uint8{0x20, 0x1E, 0x01, 0x00, 0x75, 0x00, 0x80, 0xFB},
			tx:          []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x75},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Ok",
			desc:        "Invalid checksum in the first measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x1E, 0x01, 0x00, 0x75, 0x00, 0x80, 0x00},
			tx:          []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x75},
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Ok__TVOC_256_CRC_Error",
			desc:        "Invalid checksum in the second measurement block. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x1E, 0x01, 0x00, 0x00, 0x00, 0x80, 0xFB},
			tx:          []uint8{0x00, 0x80, 0xFB, 0x01, 0x00, 0x00},
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "CO2_128_CRC_Error__TVOC_256_CRC_Error",
			desc:        "Invalid checksums for both measurements. The function should return a validation error.",
			txBytes:     []uint8{0x20, 0x1E, 0x01, 0x00, 0x00, 0x00, 0x80, 0x00},
			tx:          []uint8{0x00, 0x80, 0x00, 0x01, 0x00, 0x00},
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "DataBufferTooShort",
			desc:        "Data buffer provided by the caller is too small. The function should reject it.",
			txBytes:     nil,
			tx:          []uint8{0x12, 0x34, 0x37, 0x56, 0x78},
			expectError: true,
		},
		{
			name:        "DataBufferTooLong",
			desc:        "Data buffer provided by the caller is larger than needed. The function should reject it.",
			txBytes:     nil,
			tx:          []uint8{0x12, 0x34, 0x37, 0x56, 0x78, 0x7D, 0xFF},
			expectError: true,
		},
		{
			name:        "I2C_HardwareError",
			desc:        "Simulate hardware bus error during measurement.",
			txBytes:     nil,
			tx:          nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError}
			dev := Device{I2C: &i2c}

			err := dev.SetIaqBaseline(tc.tx)

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetIaqBaseline returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, i2c.TxData)
			}
		})
	}
}

func TestSetAbsoluteHumidity(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		tx          []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "AbsoluteHumidity_0_CRC_Ok",
			desc:        "Minimum absolute humidity value with correct checksum.",
			txBytes:     []uint8{0x20, 0x61, 0x00, 0x00, 0x81},
			tx:          []uint8{0x00, 0x00, 0x81},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "AbsoluteHumidity_Max16bit_CRC_Ok",
			desc:        "Maximum absolute humidity value with correct checksum.",
			txBytes:     []uint8{0x20, 0x61, 0xFF, 0xFF, 0xAC},
			tx:          []uint8{0xFF, 0xFF, 0xAC},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "AbsoluteHumidity_Shift_CRC_Ok",
			desc:        "Different non-zero values for most and least significant bytes to ensure the driver does not mix them up.",
			txBytes:     []uint8{0x20, 0x61, 0x12, 0x34, 0x37},
			tx:          []uint8{0x12, 0x34, 0x37},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "AbsoluteHumidity_128_CRC_Ok",
			desc:        "Valid standard absolute humidity reading with correct checksum.",
			txBytes:     []uint8{0x20, 0x61, 0x00, 0x80, 0xFB},
			tx:          []uint8{0x00, 0x80, 0xFB},
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "AbsoluteHumidity_128_CRC_Error",
			desc:        "Invalid checksum. The function should return a validation error and prevent hardware transmission.",
			txBytes:     nil,
			tx:          []uint8{0x00, 0x80, 0x00},
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "DataBufferTooShort",
			desc:        "Data buffer provided by the caller is too small. The function should reject it.",
			txBytes:     nil,
			tx:          []uint8{0x12, 0x34},
			expectError: true,
		},
		{
			name:        "DataBufferTooLong",
			desc:        "Data buffer provided by the caller is larger than needed. The function should reject it.",
			txBytes:     nil,
			tx:          []uint8{0x12, 0x34, 0x56, 0x78},
			expectError: true,
		},
		{
			name:        "I2C_HardwareError",
			desc:        "Simulate hardware bus error during measurement.",
			txBytes:     nil,
			tx:          nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError}
			dev := Device{I2C: &i2c}

			err := dev.SetAbsoluteHumidity(tc.tx)

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetAbsoluteHumidity returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, i2c.TxData)
			}
		})
	}
}

func TestMeasureTest(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		txBytes     []uint8
		rxBytes     []uint8
		rx          []uint8
		i2cError    error
		expectError bool
	}{
		{
			name:        "Test_Ok_CRC_Ok",
			desc:        "Valid self-test response indicating hardware is fully functional, with correct checksum.",
			txBytes:     []uint8{0x20, 0x32},
			rxBytes:     []uint8{0xD4, 0x00, 0xC6},
			rx:          make([]uint8, 3),
			i2cError:    nil,
			expectError: false,
		},
		{
			name:        "Test_Error_CRC_Ok",
			desc:        "Self-test response indicating hardware failure, but transmission and checksum are mathematically correct.",
			txBytes:     []uint8{0x20, 0x32},
			rxBytes:     []uint8{0xD4, 0xFF, 0x6A},
			rx:          make([]uint8, 3),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "Test_Ok_CRC_Error",
			desc:        "Self-test response has an invalid checksum. The function should reject it.",
			txBytes:     []uint8{0x20, 0x32},
			rxBytes:     []uint8{0xD4, 0x00, 0xFF},
			rx:          make([]uint8, 3),
			i2cError:    nil,
			expectError: true,
		},
		{
			name:        "DataBufferTooShort",
			desc:        "Data buffer provided by the caller is too small. The function should reject it.",
			txBytes:     []uint8{0x20, 0x32},
			rxBytes:     []uint8{0xD4, 0x00, 0xC6},
			rx:          make([]uint8, 2),
			expectError: true,
		},
		{
			name:        "DataBufferTooLong",
			desc:        "Data buffer provided by the caller is larger than needed. The function should reject it.",
			txBytes:     []uint8{0x20, 0x32},
			rxBytes:     []uint8{0xD4, 0x00, 0xC6},
			rx:          make([]uint8, 4),
			expectError: true,
		},
		{
			name:        "I2C_HardwareError",
			desc:        "Simulate hardware bus error during measurement.",
			txBytes:     nil,
			rxBytes:     nil,
			i2cError:    fmt.Errorf("I2C bus error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i2c := MockI2C{ReturnError: tc.i2cError, RxData: tc.rxBytes}
			dev := Device{I2C: &i2c}

			err := dev.MeasureTest(tc.rx)

			if tc.expectError == true {
				if err == nil {
					t.Fatalf("FAIL: %s\nExpected error but got nil", tc.desc)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nMeasureTest returned: %v", tc.desc, err)
			}

			if !bytes.Equal(i2c.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to I2C!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, i2c.TxData)
			}

			if !bytes.Equal(i2c.RxData, tc.rx) {
				t.Errorf("FAIL: %s\nWrong bytes received from I2C!\nExpected: [%# x]\nGot:      [%# x]", tc.desc, tc.rx, i2c.RxData)
			}
		})
	}
}
