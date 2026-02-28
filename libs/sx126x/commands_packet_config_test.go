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

// 13.4.2 SetPacketType
func TestSetPacketType(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		packet  PacketType
		txBytes []uint8
	}{
		{
			name:    "GFSK",
			desc:    "Verifies that the driver correctly formats the SPI command to configure the radio transceiver for Gaussian Frequency Shift Keying modulation.",
			packet:  PacketTypeGFSK,
			txBytes: []uint8{0x8A, 0x00},
		},
		{
			name:    "LoRa",
			desc:    "Verifies that the driver correctly formats the SPI command to configure the radio transceiver for Long Range modulation.",
			packet:  PacketTypeLoRa,
			txBytes: []uint8{0x8A, 0x01},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetPacketType(tc.packet)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetPacketType returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.4.3 GetPacketType
func TestGetPacketType(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		commands []uint8
		tx       []uint8
		rx       []uint8
	}{
		{
			name:     "GFSK",
			desc:     "Verifies decoding of the GFSK modulation packet type from the SPI response.",
			commands: []uint8{uint8(CmdGetPacketType), OpCodeNop, OpCodeNop},
			tx:       []uint8{0x11, 0x00, 0x00},
			rx:       []uint8{0x00, 0x01, 0x00},
		},
		{
			name:     "Lora",
			desc:     "Verifies decoding of the LoRa modulation packet type from the SPI response.",
			commands: []uint8{uint8(CmdGetPacketType), OpCodeNop, OpCodeNop},
			tx:       []uint8{0x11, 0x00, 0x00},
			rx:       []uint8{0x00, 0x01, 0x01},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			dev := Device{SPI: &spi}

			status, err := dev.GetPacketType()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetPacketType returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.tx) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.tx, spi.TxData)
			}

			if !bytes.Equal(spi.RxData, tc.rx) {
				t.Errorf("FAIL: %s\nWrong bytes read from SPI!\nExpected: [%# x]\nGot: [%# x]", tc.desc, tc.rx, spi.RxData)
			}

			switch status {
			case uint8(PacketTypeGFSK):
				return
			case uint8(PacketTypeLoRa):
				return
			default:
				t.Errorf("FAIL: %s\nWrong status returned from the SX126x modem!\nExpected: [% x] or [% x]\nGot: [% x]", tc.desc, PacketTypeGFSK, PacketTypeLoRa, status)
			}
		})
	}
}

// 13.4.5 SetModulationParams
func TestSetModulationParams(t *testing.T) {
	type lora struct {
		sf   uint8
		cr   uint8
		ldro bool
	}

	type fsk struct {
		br uint64
		ps float32
		fd uint64
	}

	tests := []struct {
		name        string
		desc        string
		modem       string
		bw          uint64
		lora        *lora
		fsk         *fsk
		options     func(d *Device) []OptionsModulation
		txBytes     []uint8
		expectError bool
	}{
		// LoRa
		{
			name:  "SF7_BW125_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly formats the SPI command with standard default LoRa parameters, including a typical spreading factor, bandwidth, and error coding rate, without low data rate optimization.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_LDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly formats the SPI command with standard default LoRa parameters while explicitly enabling low data rate optimization.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x01},
			expectError: false,
		},
		{
			name:  "SF5_BW125_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly processes the minimum allowable spreading factor alongside standard bandwidth and coding rate settings.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   5,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x05, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF12_BW125_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly processes the maximum allowable spreading factor alongside standard bandwidth and coding rate settings.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   12,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x0C, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF0_BW125_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring a below-minimum spreading factor value is safely clamped or defaulted to a standard value.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   0,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF15_BW125_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring an above-maximum spreading factor value is safely clamped or defaulted to a standard value.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   15,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW0_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring a zero bandwidth value is safely clamped or defaulted to a standard intermediate bandwidth.",
			modem: "lora",
			bw:    0,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW900_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring an excessively high bandwidth value is safely clamped or defaulted to a standard intermediate bandwidth.",
			modem: "lora",
			bw:    900000,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW7.8_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly processes the minimum valid bandwidth configuration alongside standard spreading factor and coding rate settings.",
			modem: "lora",
			bw:    7800,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x00, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW500_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly processes the maximum valid bandwidth configuration alongside standard spreading factor and coding rate settings.",
			modem: "lora",
			bw:    500000,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x06, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR0_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring a below-minimum coding rate value is safely clamped or defaulted to a standard rate.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   0,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR8_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly processes the maximum valid coding rate configuration alongside standard spreading factor and bandwidth settings.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   8,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x04, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR4_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's fallback logic by ensuring a border-case, below-minimum coding rate value is safely clamped or defaulted to a standard rate.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   4,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF5_BW500_CR5_NoLDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly formats the SPI command for a high-speed transmission profile using the minimum spreading factor and maximum bandwidth.",
			modem: "lora",
			bw:    500000,
			lora: &lora{
				sf:   5,
				cr:   5,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x05, 0x06, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF12_BW7.8_CR8_LDRO_LoRa_Default",
			desc:  "Verifies that the driver correctly formats the SPI command for an extreme long-range, low-speed transmission profile, combining maximum spreading factor, minimum bandwidth, maximum coding rate, and enabled low data rate optimization.",
			modem: "lora",
			bw:    7800,
			lora: &lora{
				sf:   12,
				cr:   8,
				ldro: true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x0C, 0x00, 0x04, 0x01},
			expectError: false,
		},
		{
			name:  "SF0_BW0_CR0_NoLDRO_LoRa_Default",
			desc:  "Verifies the driver's comprehensive fallback logic when provided with completely zeroed, invalid parameters, ensuring all values are safely restored to workable defaults.",
			modem: "lora",
			bw:    0,
			lora: &lora{
				sf:   0,
				cr:   0,
				ldro: false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF99_BW999_CR99_LDRO_LoRa_Default",
			desc:  "Verifies the driver's comprehensive fallback logic when provided with absurdly high, invalid parameters, ensuring all values are safely restored to workable defaults while maintaining requested boolean flags.",
			modem: "lora",
			bw:    999000,
			lora: &lora{
				sf:   99,
				cr:   99,
				ldro: true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x01},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_LDRO_LoRa_ModulationConfigLoRa",
			desc:  "Verifies that the multi-parameter configuration option correctly overrides all base LoRa modulation settings and formats the SPI payload accordingly.",
			modem: "lora",
			bw:    0,
			lora: &lora{
				sf:   0,
				cr:   0,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationConfigLoRa(7, 5, 125000*physic.Hertz, true)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x01},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_NoLDRO_LoRa_ModulationBW",
			desc:  "Verifies that the individual bandwidth configuration option correctly updates the base parameter and packs the proper filter index into the payload.",
			modem: "lora",
			bw:    0,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationBW(125000 * physic.Hertz)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_NoLDRO_LoRa_ModulationCR",
			desc:  "Verifies that the individual coding rate configuration option correctly overrides the base setting without affecting the other parameters.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   0,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationCR(5)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_NoLDRO_LoRa_ModulationSF",
			desc:  "Verifies that the individual spreading factor configuration option correctly overrides the base setting without affecting the other parameters.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   0,
				cr:   5,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationSF(7)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_LDRO_LoRa_ModulationLDRO",
			desc:  "Verifies that the individual low data rate optimization configuration option correctly toggles the corresponding flag in the SPI payload.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   7,
				cr:   5,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationLDRO(true)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x01},
			expectError: false,
		},
		{
			name:  "SF7_BW125_CR5_NoLDRO_LoRa_ModulationSF_ModulationCR",
			desc:  "Verifies that multiple individual functional options can be chained together to sequentially override specific base LoRa parameters before payload construction.",
			modem: "lora",
			bw:    125000,
			lora: &lora{
				sf:   0,
				cr:   0,
				ldro: false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationSF(7), d.ModulationCR(5)}
			},
			txBytes:     []uint8{0x8B, 0x07, 0x04, 0x01, 0x00},
			expectError: false,
		},

		// FSK
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies that the driver correctly formats the SPI command for a standard frequency shift keying profile using typical bit rate, frequency deviation, Gaussian pulse shape, and bandwidth parameters.",
			modem: "fsk",
			bw:    9700, // ~2*(fdev+(br/2))
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR0.6_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies that the driver correctly processes the absolute minimum allowed bit rate configuration without payload corruption or mathematical error.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 600,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x1A, 0x0A, 0xAA, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR300.0_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies that the driver correctly processes the absolute maximum standard bit rate configuration for frequency shift keying modulation.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 300000,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x00, 0x0D, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BRMax24bit_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies the mathematical stability and bitwise packing logic when the bit rate calculation formula processes the maximum possible twenty-four-bit unsigned integer value.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0xFFFFFF,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BROverflow_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies the mathematical stability and bitwise packing logic when the bit rate calculation formula processes a value exceeding the twenty-four-bit limit, checking for proper upper-byte truncation.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0xFF000000,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BRMinNonZero_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies that the bit rate calculation correctly processes the smallest non-zero unsigned integer input without causing boundary or shifting errors.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0x000001,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BRMsbOnly_PS0.5_BW9.7_FD2.4_FSK_Default",
			desc:  "Verifies the bit rate calculation when provided with an input where only the most significant bit of the twenty-four-bit field is active.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0x800000,
				ps: 0.5,
				fd: 2400,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD0_FSK_Default",
			desc:  "Verifies that the driver correctly formats the payload when the frequency deviation parameter is entirely zeroed, ensuring proper packing of an unmodulated carrier state.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 0,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FDMax24bit_FSK_Default",
			desc:  "Verifies the driver's handling of the theoretical maximum twenty-four-bit unsigned integer value for frequency deviation without logic panics.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 0xFFFFFF,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x0C, 0x6F, 0x78},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FDMinNonZero_FSK_Default",
			desc:  "Verifies the frequency deviation mathematical conversion correctly packs the lowest possible discrete fractional step into the payload.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 0x000001,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x00, 0x01},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FDMsbOnly_FSK_Default",
			desc:  "Verifies the frequency deviation conversion logic using a boundary input containing only the highest active bit within the designated bit range.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 0x800000,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x86, 0x37, 0xBD},
			expectError: false,
		},
		{
			name:  "BR0.6_PS0_BW4.8_FD0_FSK_Default",
			desc:  "Verifies the payload construction for a raw frequency shift keying configuration utilizing minimum bit rate and bandwidth settings with the pulse shaping filter fully disabled.",
			modem: "fsk",
			bw:    4800,
			lora:  nil,
			fsk: &fsk{
				br: 600,
				ps: 0.0,
				fd: 0,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x1A, 0x0A, 0xAA, 0x00, 0x1F, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:  "BR300k_PS1_BW467_FD83_FSK_Default",
			desc:  "Verifies that the driver correctly formats the SPI command for a high-speed GFSK profile using a wide receiver bandwidth, maximum standard bit rate, broad frequency deviation, and a less restrictive Gaussian pulse shape.",
			modem: "fsk",
			bw:    467000,
			lora:  nil,
			fsk: &fsk{
				br: 300000,
				ps: 1.0,
				fd: 83500,
			},
			options:     nil,
			txBytes:     []uint8{0x8B, 0x00, 0x0D, 0x55, 0x0B, 0x09, 0x01, 0x56, 0x04},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_ModulationConfigFSK",
			desc:  "Verifies that the multi-parameter configuration option correctly overrides all base FSK modulation settings, performs necessary mathematical conversions, and formats the SPI payload accordingly.",
			modem: "fsk",
			bw:    0,
			lora:  nil,
			fsk: &fsk{
				br: 0,
				ps: 0.0,
				fd: 0,
			},
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationConfigFSK(4800, 2400, 9700*physic.Hertz, 0.5)}
			},
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_ModulationBR",
			desc:  "Verifies that the individual bit rate configuration option correctly overrides the base setting and computes the correct PLL step values without affecting the other parameters.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0,
				ps: 0.5,
				fd: 2400,
			},
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationBR(4800)}
			},
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_ModulationPS",
			desc:  "Verifies that the individual pulse shape configuration option correctly updates the base parameter and packs the proper Gaussian filter index into the payload.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.0,
				fd: 2400,
			},
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationPS(0.5)}
			},
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_ModulationFD",
			desc:  "Verifies that the individual frequency deviation configuration option correctly overrides the base setting and accurately calculates the required fractional register values.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 4800,
				ps: 0.5,
				fd: 0,
			},
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationFD(2400)}
			},
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},
		{
			name:  "BR4.8_PS0.5_BW9.7_FD2.4_FSK_ModulationBR_ModulationPS",
			desc:  "Verifies that the individual frequency deviation configuration option correctly overrides the base setting and accurately calculates the required fractional register values.",
			modem: "fsk",
			bw:    9700,
			lora:  nil,
			fsk: &fsk{
				br: 0,
				ps: 0,
				fd: 2400,
			},
			options: func(d *Device) []OptionsModulation {
				return []OptionsModulation{d.ModulationBR(4800), d.ModulationPS(0.5)}
			},
			txBytes:     []uint8{0x8B, 0x03, 0x41, 0x55, 0x09, 0x1E, 0x00, 0x09, 0xD4},
			expectError: false,
		},

		// Misc
		{
			name:        "UnknownModem",
			desc:        "Verifies that the driver safely aborts and returns an error when attempting to set modulation parameters for an unsupported or uninitialized modem type.",
			modem:       "invalid",
			bw:          0,
			lora:        nil,
			fsk:         nil,
			options:     nil,
			txBytes:     nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			cfg := Config{
				Modem:     tc.modem,
				Bandwidth: tc.bw,
				// Workarounds: &Workarounds{},  // TODO: Add 15.1 Modulation Quality with 500 kHz LoRa Bandwidth workaround to tests
			}

			if tc.lora != nil {
				cfg.LoRa = &LoRa{
					SpreadingFactor: tc.lora.sf,
					CodingRate:      tc.lora.cr,
					LDRO:            tc.lora.ldro,
				}
			}

			if tc.fsk != nil {
				cfg.FSK = &FSK{
					Bitrate:            tc.fsk.br,
					PulseShape:         tc.fsk.ps,
					FrequencyDeviation: tc.fsk.fd,
				}
			}

			dev := Device{SPI: &spi, Config: &cfg}

			var opts []OptionsModulation
			if tc.options != nil {
				opts = tc.options(&dev)
			}
			err := dev.SetModulationParams(opts...)

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetModulationParams returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

func TestSetPacketParams(t *testing.T) {
	type lora struct {
		headerImplicit bool
		invertedIQ     bool
		crc            bool
	}

	type fsk struct {
		preDetector FskPreambleDetector
		syncWord    FskSyncWord
		addrComp    uint8
		packetType  string
		crc         string
		whitening   bool
	}

	tests := []struct {
		name        string
		desc        string
		modem       string
		preamble    uint16
		payload     uint8
		lora        *lora
		fsk         *fsk
		options     func(d *Device) []OptionsPacket
		txBytes     []uint8
		expectError bool
	}{
		// LoRa
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_Default",
			desc:     "Verifies that the driver correctly formats the SPI command with standard LoRa packet parameters, including an explicit header, enabled CRC, and inverted IQ setup.",
			modem:    "lora",
			preamble: 12,
			payload:  32,
			lora: &lora{
				headerImplicit: false,
				crc:            true,
				invertedIQ:     true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble0_HeaderImplicit_Payload0_CRCOff_IQStandard_LoRa_Default",
			desc:     "Verifies the packing logic for the absolute minimum allowable LoRa packet parameters, disabling all optional payload features like CRC and explicit headers.",
			modem:    "lora",
			preamble: 0,
			payload:  0,
			lora: &lora{
				headerImplicit: true,
				crc:            false,
				invertedIQ:     false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:     "PreambleMax16bit_HeaderImplicit_PayloadMax8bit_CRCOn_IQInverted_LoRa_Default",
			desc:     "Verifies the packing logic and boundary handling when providing the theoretical maximum unsigned integer values for preamble and payload lengths.",
			modem:    "lora",
			preamble: 0xFFFF,
			payload:  0xFF,
			lora: &lora{
				headerImplicit: true,
				crc:            true,
				invertedIQ:     true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8C, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "PreambleShift_HeaderImplicit_PayloadShift_CRCOn_IQInverted_LoRa_Default",
			desc:     "Verifies the bitwise shift operations for multibyte packet parameters using a recognizable hex pattern to ensure upper and lower bytes are not swapped or truncated.",
			modem:    "lora",
			preamble: 0x1234,
			payload:  0x56,
			lora: &lora{
				headerImplicit: true,
				crc:            true,
				invertedIQ:     true,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8C, 0x12, 0x34, 0x01, 0x56, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble8_HeaderExplicit_Payload64_CrcOff_IQStandard_LoRa_Default",
			desc:     "Verifies a specific cross-combination of boolean packet flags using an explicit header alongside standard IQ and disabled CRC.",
			modem:    "lora",
			preamble: 8,
			payload:  64,
			lora: &lora{
				headerImplicit: false,
				crc:            false,
				invertedIQ:     false,
			},
			fsk:         nil,
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x08, 0x00, 0x40, 0x00, 0x00},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketLoRaConfig",
			desc:     "",
			modem:    "lora",
			preamble: 0xFF,
			payload:  0xFF,
			lora: &lora{
				headerImplicit: true,
				crc:            false,
				invertedIQ:     false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketLoRaConfig(12, HeaderExplicit, 32, CrcOn, IqInverted)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketPreLen",
			desc:     "Verifies that the individual preamble length configuration option correctly overrides the base setting without affecting the payload length or header flags.",
			modem:    "lora",
			preamble: 0,
			payload:  32,
			lora: &lora{
				headerImplicit: false,
				crc:            true,
				invertedIQ:     true,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketPreLen(12)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketHT",
			desc:     "Verifies that the individual header type configuration option correctly overrides the base setting, toggling the explicit header flag in the SPI payload.",
			modem:    "lora",
			preamble: 12,
			payload:  32,
			lora: &lora{
				headerImplicit: true,
				crc:            true,
				invertedIQ:     true,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketHT(HeaderExplicit)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketPayLen",
			desc:     "Verifies that the individual payload length configuration option correctly updates the base parameter and packs the proper byte size into the payload.",
			modem:    "lora",
			preamble: 12,
			payload:  0,
			lora: &lora{
				headerImplicit: false,
				crc:            true,
				invertedIQ:     true,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketPayLen(32)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketIQ",
			desc:     "Verifies that the individual IQ polarity configuration option correctly toggles the corresponding boolean flag in the SPI payload.",
			modem:    "lora",
			preamble: 12,
			payload:  32,
			lora: &lora{
				headerImplicit: false,
				crc:            true,
				invertedIQ:     false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketIQ(IqInverted)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble12_HeaderExplicit_Payload32_CRCOn_IQInverted_LoRa_PacketHT_PacketIQ",
			desc:     "Verifies that multiple individual functional options can be chained together to sequentially override specific base LoRa packet parameters before payload construction.",
			modem:    "lora",
			preamble: 12,
			payload:  32,
			lora: &lora{
				headerImplicit: true,
				crc:            true,
				invertedIQ:     false,
			},
			fsk: nil,
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketHT(HeaderExplicit), d.PacketIQ(IqInverted)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x0C, 0x00, 0x20, 0x01, 0x01},
			expectError: false,
		},

		// FSK
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_Default",
			desc:     "Verifies that the driver correctly formats the SPI command with standard industrial FSK packet parameters, establishing the baseline execution path for variable length configuration.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble0_PreDetector0_SyncWord0_AddrCompOff_PacketStatic_Payload0_CrcOff_WhiteningOff_FSK_Default",
			desc:     "Verifies the packing logic for the minimum allowable FSK packet parameters, disabling all optional hardware features like CRC, whitening, and setting lengths to zero.",
			modem:    "fsk",
			preamble: 0,
			payload:  0,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLenOff,
				syncWord:    FskSyncWordLength0,
				addrComp:    0,
				packetType:  "static",
				crc:         "0",
				whitening:   false,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc1Inverted_WhiteningOn_FSK_Default",
			desc:     "Verifies the parsing logic and byte construction when selecting the 1-byte inverted CRC variant for FSK modulation.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "1_inv",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x04, 0x01},
			expectError: false,
		},
		{
			name:     "PreambleMax16bit_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_PayloadMax8bit_Crc2Standard_WhiteningOn_FSK_Default",
			desc:     "Verifies the packing logic and boundary handling when providing the theoretical maximum unsigned integer values for preamble and payload lengths in FSK mode.",
			modem:    "fsk",
			preamble: 0xFFFF,
			payload:  0xFF,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0xFF, 0xFF, 0x05, 0x10, 0x00, 0x01, 0xFF, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "PreambleShift_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_PayloadShift_Crc2Standard_WhiteningOn_FSK_Default",
			desc:     "Verifies the bitwise shift operations for multibyte FSK packet parameters using a recognizable hex pattern to ensure upper and lower bytes are not swapped or truncated.",
			modem:    "fsk",
			preamble: 0x1234,
			payload:  0x56,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x12, 0x34, 0x05, 0x10, 0x00, 0x01, 0x56, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompNode_PacketStatic_Payload32_Crc1Standard_WhiteningOff_FSK_Default",
			desc:     "Verifies a specific cross-combination of FSK packet flags, including node address comparison enabled alongside static payload length and standard 1-byte CRC.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    1,
				packetType:  "static",
				crc:         "1",
				whitening:   false,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x01, 0x00, 0x20, 0x00, 0x00},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompAll_PacketVariable_Payload32_Crc2Inverted_WhiteningOn_FSK_Default",
			desc:     "Verifies a specific cross-combination of FSK packet flags, including broadcast address comparison enabled alongside variable payload length and 2-byte inverted CRC.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    2,
				packetType:  "variable",
				crc:         "2_inv",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x02, 0x01, 0x20, 0x06, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompAll_PacketInvalid_Payload32_Crc2Invalid_WhiteningOn_FSK_Default",
			desc:     "Verifies the fallback logic and safe state recovery of the driver when provided with non-existent string variants for packet type and CRC in the configuration struct.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    2,
				packetType:  "invalid",
				crc:         "invalid",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x02, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetectorInvalid_SyncWordInvalid_AddrCompInvalid_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_Default",
			desc:     "Verifies the fallback logic and safe state recovery of the driver when provided with hardware values that are missing from the internal validating enumerations.",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: FskPreambleDetector(0xFF),
				syncWord:    FskSyncWord(0xFF),
				addrComp:    0xFF,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options:     nil,
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketPreLen",
			desc:     "",
			modem:    "fsk",
			preamble: 0,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketPreLen(32)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketFskCRC",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "0",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketFskCRC(CRC2)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketFskCRC_PacketPreDet",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: 0,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "0",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketFskCRC(CRC2), d.PacketPreDet(PreambleDetLen16)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketPreDet",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: 0,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketPreDet(PreambleDetLen16)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketFskSW",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    0,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketFskSW(FskSyncWordLength2)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketAddrCmp",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0xFF,
				packetType:  "variable",
				crc:         "2",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketAddrCmp(AddrCompOff)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketFskType",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "static",
				crc:         "2",
				whitening:   true,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketFskType(PacketTypeGFSKVariable)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketWhitening",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: PreambleDetLen16,
				syncWord:    FskSyncWordLength2,
				addrComp:    0,
				packetType:  "variable",
				crc:         "2",
				whitening:   false,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketWhitening(WhiteningOn)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},
		{
			name:     "Preamble32_PreDetector16_SyncWord2_AddrCompOff_PacketVariable_Payload32_Crc2Standard_WhiteningOn_FSK_PacketFskConfig",
			desc:     "",
			modem:    "fsk",
			preamble: 32,
			payload:  32,
			lora:     nil,
			fsk: &fsk{
				preDetector: 0xFF,
				syncWord:    0xFF,
				addrComp:    0xFF,
				packetType:  "",
				crc:         "",
				whitening:   false,
			},
			options: func(d *Device) []OptionsPacket {
				return []OptionsPacket{d.PacketFskConfig(PreambleDetLen16, FskSyncWordLength2, AddrCompOff, PacketTypeGFSKVariable, CRC2, WhiteningOn)}
			},
			txBytes:     []uint8{0x8C, 0x00, 0x20, 0x05, 0x10, 0x00, 0x01, 0x20, 0x02, 0x01},
			expectError: false,
		},

		// Misc
		{
			name:        "UnknownModem",
			desc:        "",
			modem:       "invalid",
			preamble:    0,
			payload:     0,
			lora:        nil,
			fsk:         nil,
			options:     nil,
			txBytes:     nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			cfg := Config{
				Modem: tc.modem,
				// Workarounds: &Workarounds{},	// TODO: Add 15.4 Optimizing the Inverted IQ Operation workaround tests
			}

			if tc.lora != nil {
				cfg.PreambleLength = tc.preamble
				cfg.PayloadLength = tc.payload
				cfg.LoRa = &LoRa{
					HeaderImplicit: tc.lora.headerImplicit,
					InvertedIQ:     tc.lora.invertedIQ,
					CRC:            tc.lora.crc,
				}
			}

			if tc.fsk != nil {
				cfg.PreambleLength = tc.preamble
				cfg.PayloadLength = tc.payload
				cfg.FSK = &FSK{
					PreambleDetectionLength: uint8(tc.fsk.preDetector),
					SyncWordDetectionLength: uint8(tc.fsk.syncWord),
					AddressComparison:       tc.fsk.addrComp,
					PacketType:              tc.fsk.packetType,
					CRC:                     tc.fsk.crc,
					Whitening:               tc.fsk.whitening,
				}
			}

			dev := Device{SPI: &spi, Config: &cfg}

			var opts []OptionsPacket
			if tc.options != nil {
				opts = tc.options(&dev)
			}
			err := dev.SetPacketParams(opts...)

			if tc.expectError == true {
				if err == nil {
					t.Errorf("FAIL: %s\nExpected an error for modem %q, but got nil", tc.desc, tc.modem)
				}
				return
			}

			if err != nil {
				t.Fatalf("FAIL: %s\nSetPacketParams returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

func TestSetCadParams(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		params  *ConfigCAD
		options func(d *Device) []OptionsCAD
		txBytes []uint8
	}{
		{
			name: "SymbolNum2_DetectPeak20_DetectMin10_ExitRX_Timeout100_Default",
			desc: "Verifies that the driver correctly formats the SPI command with standard CAD parameters, utilizing 2 symbols for detection and automatically entering RX mode upon success.",
			params: &ConfigCAD{
				SymbolNumber:     2,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          100,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum0_DetectPeak20_DetectMin10_ExitRX_Timeout100_Default",
			desc: "Verifies the driver's fallback logic when provided with an invalid symbol number, ensuring it safely clamps to a standard default value (e.g., 2 symbols).",
			params: &ConfigCAD{
				SymbolNumber:     0,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          100,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum2_DetectPeak20_DetectMin10_ExitInvalid_Timeout100_Default",
			desc: "Verifies the driver's fallback logic when provided with an invalid exit mode, ensuring it safely defaults to entering RX mode.",
			params: &ConfigCAD{
				SymbolNumber:     2,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         0xFF,
				Timeout:          100,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum0_DetectPeak0_DetectMin0_ExitInvalid_Timeout0_Default",
			desc: "Verifies the comprehensive fallback logic when provided with entirely zeroed or invalid boundary parameters, ensuring safe driver operation and payload construction.",
			params: &ConfigCAD{
				SymbolNumber:     0,
				DetectionPeak:    0,
				DetectionMinimum: 0,
				ExitMode:         0xFF,
				Timeout:          0,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name: "SymbolNumMax8bit_DetectPeakMax8bit_DetectMinMax8bit_ExitMax8bit_TimeoutMax24bit_Default",
			desc: "Verifies the mathematical stability and bitwise packing logic when all CAD parameters are set to their maximum theoretical bounds.",
			params: &ConfigCAD{
				SymbolNumber:     0xFF,
				DetectionPeak:    0xFF,
				DetectionMinimum: 0xFF,
				ExitMode:         0xFF,
				Timeout:          0xFFFFFF,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0xFF, 0xFF},
		},
		{
			name: "SymbolNum2_DetectPeak20_DetectMin10_ExitRX_TimeoutOverflow_Default",
			desc: "Verifies the driver correctly handles a timeout value exceeding the 24-bit limit, clamping it safely to the maximum allowable value (0xFFFFFF) without bitwise corruption.",
			params: &ConfigCAD{
				SymbolNumber:     2,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          0x12FFFFFF,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x14, 0x0A, 0x01, 0xFF, 0xFF, 0xFF},
		},
		{
			name: "SymbolNum2_DetectPeak20_DetectMin10_ExitStandby_TimeoutShift_Default",
			desc: "Verifies the multibyte bitwise shift logic for the 24-bit timeout parameter using a recognizable hex pattern to ensure accurate byte ordering.",
			params: &ConfigCAD{
				SymbolNumber:     2,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         0,
				Timeout:          0x123456,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x01, 0x14, 0x0A, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name: "SymbolNum16_DetectPeak20_DetectMin10_ExitRX_Timeout100_Default",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     16,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         0,
				Timeout:          0x123456,
			},
			options: nil,
			txBytes: []uint8{0x88, 0x04, 0x14, 0x0A, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADSym",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     0,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          100,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADSym(CadOn4Symb)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADConfig",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     0,
				DetectionPeak:    0,
				DetectionMinimum: 0,
				ExitMode:         0,
				Timeout:          0,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADConfig(CadOn4Symb, 20, 10, CadExitRx, 100)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADPeak",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     4,
				DetectionPeak:    0,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          100,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADPeak(20)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADMin",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     4,
				DetectionPeak:    20,
				DetectionMinimum: 0,
				ExitMode:         1,
				Timeout:          100,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADMin(10)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADExit",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     4,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         0,
				Timeout:          100,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADExit(CadExitRx)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum4_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADTimeout",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     4,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          0,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADTimeout(100)}
			},
			txBytes: []uint8{0x88, 0x02, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
		{
			name: "SymbolNum16_DetectPeak20_DetectMin10_ExitRX_Timeout100_CADTimeout_CADSym",
			desc: "",
			params: &ConfigCAD{
				SymbolNumber:     0,
				DetectionPeak:    20,
				DetectionMinimum: 10,
				ExitMode:         1,
				Timeout:          0,
			},
			options: func(d *Device) []OptionsCAD {
				return []OptionsCAD{d.CADTimeout(100), d.CADSym(CadOn16Symb)}
			},
			txBytes: []uint8{0x88, 0x04, 0x14, 0x0A, 0x01, 0x00, 0x00, 0x64},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			cfg := Config{
				LoRa: &LoRa{
					CAD: &CAD{
						SymbolNumber:     tc.params.SymbolNumber,
						DetectionPeak:    tc.params.DetectionPeak,
						DetectionMinimum: tc.params.DetectionMinimum,
						ExitMode:         tc.params.ExitMode,
						Timeout:          tc.params.Timeout,
					},
				},
			}
			dev := Device{SPI: &spi, Config: &cfg}

			var opts []OptionsCAD
			if tc.options != nil {
				opts = tc.options(&dev)
			}
			err := dev.SetCadParams(opts...)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetCadParams returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}
