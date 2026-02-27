package sx126x

import (
	"bytes"
	"testing"
)

// 13.5.5 GetStats
func TestGetStats(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		modem   string
		stats   PacketStats
		tx      []uint8
		rx      []uint8
		txBytes []uint8
	}{
		// LoRa
		{
			name:  "NbPkt0_NbPktCrc0_NbHdrErr0_LoRa",
			desc:  "Verifies decoding when no packets have been received and all counters are clear.",
			modem: "lora",
			stats: PacketStats{
				TotalReceived: 0,
				CrcErrors:     0,
				HeaderErrors:  0,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPktMax16bit_NbPktCrcMax16bit_NbHdrErrMax16bit_LoRa",
			desc:  "Verifies decoding at the maximum boundary for all statistic counters.",
			modem: "lora",
			stats: PacketStats{
				TotalReceived: 0xFFFF,
				CrcErrors:     0xFFFF,
				HeaderErrors:  0xFFFF,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPktShift_NbPktCrcShift_NbHdrErrShift_LoRa",
			desc:  "Verifies correct MSB and LSB byte assembly for all statistic counters to prevent endianness issues.",
			modem: "lora",
			stats: PacketStats{
				TotalReceived: 0x1234,
				CrcErrors:     0x5678,
				HeaderErrors:  0xABCD,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPkt32_NbPktCrc0_NbHdrErr0_LoRa",
			desc:  "Verifies decoding when no packets have been received and all counters are clear.",
			modem: "lora",
			stats: PacketStats{
				TotalReceived: 32,
				CrcErrors:     0,
				HeaderErrors:  0,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},

		// FSK
		{
			name:  "NbPkt0_NbPktCrc0_NbLenErr0_FSK",
			desc:  "Verifies decoding when no packets have been received and all counters are clear.",
			modem: "fsk",
			stats: PacketStats{
				TotalReceived: 0,
				CrcErrors:     0,
				LengthErrors:  0,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPktMax16bit_NbPktCrcMax16bit_NbLenErrMax16bit_FSK",
			desc:  "Verifies decoding at the maximum boundary for all statistic counters.",
			modem: "fsk",
			stats: PacketStats{
				TotalReceived: 0xFFFF,
				CrcErrors:     0xFFFF,
				LengthErrors:  0xFFFF,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPktShift_NbPktCrcShift_NbLenErrShift_FSK",
			desc:  "Verifies correct MSB and LSB byte assembly for all statistic counters to prevent endianness issues.",
			modem: "fsk",
			stats: PacketStats{
				TotalReceived: 0x1234,
				CrcErrors:     0x5678,
				LengthErrors:  0xABCD,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "NbPkt32_NbPktCrc0_NbLenErr0_FSK",
			desc:  "Verifies decoding when no packets have been received and all counters are clear.",
			modem: "fsk",
			stats: PacketStats{
				TotalReceived: 32,
				CrcErrors:     0,
				LengthErrors:  0,
			},
			tx:      []uint8{uint8(CmdGetStats), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00},
			txBytes: []uint8{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			cfg := Config{Modem: tc.modem}
			dev := Device{SPI: &spi, Config: &cfg}

			stats, err := dev.GetStats()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetStats returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}

			if stats != tc.stats {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: [%#v]\nGot:      [%#v]", tc.desc, tc.stats, stats)
			}
		})
	}
}

// 13.5.6 ResetStats
func TestResetStats(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		clear   bool
		txBytes []uint8
	}{
		{
			name:    "InternalCacheReset",
			desc:    "",
			clear:   true,
			txBytes: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:    "NoInternalCacheReset",
			desc:    "",
			clear:   false,
			txBytes: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.ResetStats(tc.clear)

			if err != nil {
				t.Fatalf("FAIL: %s\nGetStats returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}
