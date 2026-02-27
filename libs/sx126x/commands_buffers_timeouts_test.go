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

// 13.4.8 SetBufferBaseAddress
func TestSetBufferBaseAddress(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		txBase  uint8
		rxBase  uint8
		txBytes []uint8
	}{
		{
			name:    "TxBase0_RxBase0",
			desc:    "Verifies payload construction for overlapping TX and RX buffer pointers at the zero index.",
			txBase:  0,
			rxBase:  0,
			txBytes: []uint8{0x8F, 0x00, 0x00},
		},
		{
			name:    "TxBaseMax8bit_RxBaseMax8bit",
			desc:    "Verifies payload construction at the maximum 8-bit address boundary (0xFF) for both TX and RX buffers.",
			txBase:  0xFF,
			rxBase:  0xFF,
			txBytes: []uint8{0x8F, 0xFF, 0xFF},
		},
		{
			name:    "TxBase0_RxBaseMax8bit",
			desc:    "Verifies payload construction using mixed minimum and maximum boundaries to ensure independent byte packing.",
			txBase:  0,
			rxBase:  0xFF,
			txBytes: []uint8{0x8F, 0x00, 0xFF},
		},
		{
			name:    "TxBaseMax8bit_RxBase0",
			desc:    "Verifies payload construction using mixed minimum and maximum boundaries to ensure independent byte packing.",
			txBase:  0xFF,
			rxBase:  0,
			txBytes: []uint8{0x8F, 0xFF, 0x00},
		},
		{
			name:    "TxBaseShift_RxBaseShift",
			desc:    "Verifies correct byte ordering in the SPI payload using unique hex patterns to ensure TX and RX addresses are not swapped.",
			txBase:  0x12,
			rxBase:  0x34,
			txBytes: []uint8{0x8F, 0x12, 0x34},
		},
		{
			name:    "TxBase0_RxBase128",
			desc:    "Verifies payload construction for the most common production scenario: evenly split 256-byte buffer.",
			txBase:  0,
			rxBase:  128,
			txBytes: []uint8{0x8F, 0x00, 0x80},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetBufferBaseAddress(tc.txBase, tc.rxBase)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetBufferBaseAddress returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.4.9 SetLoRaSymbNumTimeout
func TestSetLoRaSymbNumTimeout(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		number  uint8
		txBytes []uint8
	}{
		{
			name:    "SymbNumNoTimeout",
			desc:    "Verifies payload construction when disabling the LoRa symbol timeout.",
			number:  0,
			txBytes: []uint8{0xA0, 0x00},
		},
		{
			name:    "SymbNumMax8bit",
			desc:    "Verifies payload construction at the maximum 8-bit boundary for the symbol timeout.",
			number:  0xFF,
			txBytes: []uint8{0xA0, 0xFF},
		},
		{
			name:    "SymbNum4",
			desc:    "Verifies payload construction for a typical LoRa symbol timeout value.",
			number:  4,
			txBytes: []uint8{0xA0, 0x04},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{}
			dev := Device{SPI: &spi}

			err := dev.SetLoRaSymbNumTimeout(tc.number)

			if err != nil {
				t.Fatalf("FAIL: %s\nSetLoRaSymbNumTimeout returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.5.2 GetRxBufferStatus
func TestGetRxBufferStatus(t *testing.T) {
	tests := []struct {
		name       string
		desc       string
		tx         []uint8
		rx         []uint8
		expectedTx []uint8
	}{
		{
			name:       "Payload32_Address0",
			desc:       "Verifies decoding of a standard 32-byte payload length starting at buffer index 0.",
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x20, 0x00},
			expectedTx: []uint8{0x13, 0x00, 0x00, 0x00},
		},
		{
			name:       "Payload0_Address128",
			desc:       "Verifies decoding of an empty payload starting at the midpoint of the RX buffer.",
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x80},
			expectedTx: []uint8{0x13, 0x00, 0x00, 0x00},
		},
		{
			name:       "PayloadMax8bit_AddressMax8bit",
			desc:       "Verifies decoding at the maximum 8-bit boundary for both payload length and start pointer.",
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0xFF, 0xFF},
			expectedTx: []uint8{0x13, 0x00, 0x00, 0x00},
		},
		{
			name:       "Payload0_Address0",
			desc:       "Verifies decoding at the absolute minimum boundary for both payload length and start pointer.",
			tx:         []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			rx:         []uint8{0x00, 0x01, 0x00, 0x00},
			expectedTx: []uint8{0x13, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			dev := Device{SPI: &spi}

			status, err := dev.GetRxBufferStatus()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetRxBufferStatus returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.expectedTx) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedTx, spi.TxData)
			}

			if status.RXPayloadLength != tc.rx[2] {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: [%# x]\nGot:      [%# x]", tc.desc, tc.rx[2], status.RXPayloadLength)
			}

			if status.RXStartPointer != tc.rx[3] {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: [%# x]\nGot:      [%# x]", tc.desc, tc.rx[3], status.RXStartPointer)
			}
		})
	}
}
