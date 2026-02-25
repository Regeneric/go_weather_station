package sx126x

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"periph.io/x/conn/v3/physic"
)

func init() {
	discardHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(discardHandler))
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.4.1 SetRfFrequency
func TestSetRfFrequency(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		freq          physic.Frequency
		expectedBytes []uint8
	}{
		{
			name:          "RfZero",
			desc:          "Verifies that providing a zero frequency correctly evaluates through the conversion formula and results in a completely zeroed Phase-Locked Loop step payload.",
			freq:          0x00000000 * physic.Hertz,
			expectedBytes: []uint8{0x86, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:          "RfMax32bit",
			desc:          "Verifies the driver's mathematical stability when provided with the absolute maximum thirty-two-bit frequency value, ensuring that internal sixty-four-bit calculations do not panic and the final result is correctly truncated and packed into the four-byte SPI payload.",
			freq:          0xFFFFFFFF * physic.Hertz,
			expectedBytes: []uint8{0x86, 0x0C, 0x6F, 0x7A, 0x0A},
		},
		{
			name:          "RfMinNonzero",
			desc:          "Verifies that the absolute minimum non-zero frequency is correctly processed by the conversion formula without rounding down to zero, ensuring the lowest possible fractional step is evaluated.",
			freq:          0x00000001 * physic.Hertz,
			expectedBytes: []uint8{0x86, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:          "RfMSBOnly",
			desc:          "Verifies the mathematical conversion and byte packing when provided with a frequency value containing only the most significant bit, ensuring numeric boundary safety during multiplication and division.",
			freq:          0x80000000 * physic.Hertz,
			expectedBytes: []uint8{0x86, 0x86, 0x37, 0xBD, 0x05},
		},
		{
			name:          "RfShift",
			desc:          "Verifies that an arbitrary, multi-byte alternating bit pattern is correctly processed by the frequency conversion formula and accurately dispersed across the entire four-byte SPI payload.",
			freq:          0x12345678 * physic.Hertz,
			expectedBytes: []uint8{0x86, 0x13, 0x16, 0xB7, 0xE4},
		},
		{
			name:          "Rf433M",
			desc:          "Verifies that the driver correctly calculates and formats the SPI payload for the standard low-band Industrial, Scientific, and Medical radio frequency.",
			freq:          433 * physic.MegaHertz,
			expectedBytes: []uint8{0x86, 0x1B, 0x10, 0x00, 0x00},
		},
		{
			name:          "Rf868M",
			desc:          "Verifies that the driver correctly calculates and formats the SPI payload for the standard European high-band Industrial, Scientific, and Medical radio frequency.",
			freq:          868 * physic.MegaHertz,
			expectedBytes: []uint8{0x86, 0x36, 0x40, 0x00, 0x00},
		},
		{
			name:          "Rf915M",
			desc:          "Verifies that the driver correctly calculates and formats the SPI payload for the standard North American high-band Industrial, Scientific, and Medical radio frequency.",
			freq:          915 * physic.MegaHertz,
			expectedBytes: []uint8{0x86, 0x39, 0x30, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetRfFrequency(tc.freq)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetRfFrequency returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}

// # 13.4.4 SetTxParams
func TestSetTxParams(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		dbm           int8
		ramp          RampTime
		expectedBytes []uint8
	}{
		{
			name:          "Ramp10usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a ten-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x00, 0x00},
		},
		{
			name:          "Ramp20usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a twenty-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp20u,
			expectedBytes: []uint8{0x8E, 0x00, 0x01},
		},
		{
			name:          "Ramp40usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a forty-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp40u,
			expectedBytes: []uint8{0x8E, 0x00, 0x02},
		},
		{
			name:          "Ramp80usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with an eighty-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp80u,
			expectedBytes: []uint8{0x8E, 0x00, 0x03},
		},
		{
			name:          "Ramp200usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a two-hundred-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp200u,
			expectedBytes: []uint8{0x8E, 0x00, 0x04},
		},
		{
			name:          "Ramp800usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with an eight-hundred-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp800u,
			expectedBytes: []uint8{0x8E, 0x00, 0x05},
		},
		{
			name:          "Ramp1700usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a one-thousand-seven-hundred-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp1700u,
			expectedBytes: []uint8{0x8E, 0x00, 0x06},
		},
		{
			name:          "Ramp3400usZeroDbm",
			desc:          "Verifies that the driver correctly formats the SPI command to set zero decibel-milliwatt output power with a three-thousand-four-hundred-microsecond power amplifier ramp time.",
			dbm:           0,
			ramp:          PaRamp3400u,
			expectedBytes: []uint8{0x8E, 0x00, 0x07},
		},
		{
			name:          "TxPowerMax8bit",
			desc:          "Verifies that the driver correctly processes the maximum possible positive signed integer value for transmission power without payload corruption.",
			dbm:           0x7F,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x7F, 0x00},
		},
		{
			name:          "TxPowerShift",
			desc:          "Verifies that an arbitrary positive power value is correctly positioned in the payload alongside a standard ramp time configuration.",
			dbm:           0x12,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x12, 0x00},
		},
		{
			name:          "TxPowerPlus1",
			desc:          "Verifies that the driver correctly processes a minimal positive one decibel-milliwatt output power setting.",
			dbm:           1,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x01, 0x00},
		},
		{
			name:          "TxPowerPlus14",
			desc:          "Verifies that the driver correctly formats the SPI command for a standard fourteen decibel-milliwatt high-power transmission setting typically used in standard configurations.",
			dbm:           14,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x0E, 0x00},
		},
		{
			name:          "TxPowerPlus22",
			desc:          "Verifies that the driver correctly formats the SPI command for the absolute maximum twenty-two decibel-milliwatt transmission power supported by the hardware.",
			dbm:           22,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0x16, 0x00},
		},
		{
			name:          "TxPowerNeg1",
			desc:          "Verifies that the driver correctly processes a negative one decibel-milliwatt power setting, ensuring proper two's complement byte encoding.",
			dbm:           -1,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0xFF, 0x00},
		},
		{
			name:          "TxPowerNeg9",
			desc:          "Verifies that the driver correctly handles a typical negative transmission power value through proper signed two's complement conversion.",
			dbm:           -9,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0xF7, 0x00},
		},
		{
			name:          "TxPowerNeg17",
			desc:          "Verifies that the driver accurately encodes the absolute lowest configurable negative transmission power limit using two's complement representation.",
			dbm:           -17,
			ramp:          PaRamp10u,
			expectedBytes: []uint8{0x8E, 0xEF, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetTxParams(tc.dbm, tc.ramp)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetTxParams returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}
