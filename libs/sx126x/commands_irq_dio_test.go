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

// 13.3.1 SetDioIrqParams
func TestSetDioIrqParams(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		modem       string
		irqMask     IrqMask
		irqMasks    []IrqMask
		txBytes     []uint8
		expectError bool
	}{
		// LoRa IRQs - single mask (IRQ + DIO1)
		{
			name:        "TxDone_Single_LoRa",
			desc:        "Verifies that TxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a completed transmission.",
			modem:       "lora",
			irqMask:     IrqTxDone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "RxDone_Single_LoRa",
			desc:        "Verifies that RxDone is globally enabled and mapped to the DIO1 pin, ensuring the host is notified when a new packet is ready in the buffer.",
			modem:       "lora",
			irqMask:     IrqRxDone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "PreambleDetected_Single_LoRa",
			desc:        "Verifies that PreambleDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the detection of a valid preamble.",
			modem:       "lora",
			irqMask:     IrqPreambleDetected,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "SyncWordValid_Single_LoRa",
			desc:        "Verifies that attempting to enable SyncWordValid while in LoRa mode correctly returns an error, as it is an FSK-specific interrupt.",
			modem:       "lora",
			irqMask:     IrqSyncWordValid,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "HeaderValid_Single_LoRa",
			desc:        "Verifies that HeaderValid is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the reception of a valid header.",
			modem:       "lora",
			irqMask:     IrqHeaderValid,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x10, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "HeaderErr_Single_LoRa",
			desc:        "Verifies that HeaderErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a corrupted packet header.",
			modem:       "lora",
			irqMask:     IrqHeaderErr,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "CrcErr_Single_LoRa",
			desc:        "Verifies that CrcErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a payload CRC mismatch.",
			modem:       "lora",
			irqMask:     IrqCrcErr,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x40, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "CadDone_Single_LoRa",
			desc:        "Verifies that CadDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal the completion of a CAD cycle.",
			modem:       "lora",
			irqMask:     IrqCadDone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x80, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "CadDetected_Single_LoRa",
			desc:        "Verifies that CadDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal activity detection in the channel.",
			modem:       "lora",
			irqMask:     IrqCadDetected,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_Single_LoRa",
			desc:        "Verifies that Timeout is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a programmed operation timeout.",
			modem:       "lora",
			irqMask:     IrqTimeout,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "All_Single_LoRa",
			desc:        "Verifies that enabling all interrupt bits (including reserved ones) is correctly rejected by the driver safety logic.",
			modem:       "lora",
			irqMask:     IrqAll,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "None_Single_LoRa",
			desc:        "Verifies that providing an empty mask correctly disables all global interrupts and clears all DIO routing.",
			modem:       "lora",
			irqMask:     IrqNone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_Single_LoRa",
			desc:        "Verifies that a standard set of LoRa interrupts (TX, RX, Timeout, and Errors) is globally enabled and correctly routed to the DIO1 pin.",
			modem:       "lora",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x02, 0x63, 0x02, 0x63, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},

		// FSK IRQs - single mask (IRQ + DIO1)
		{
			name:        "TxDone_Single_FSK",
			desc:        "Verifies that TxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:       "fsk",
			irqMask:     IrqTxDone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "RxDone_Single_FSK",
			desc:        "Verifies that RxDone is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:       "fsk",
			irqMask:     IrqRxDone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "PreambleDetected_Single_FSK",
			desc:        "Verifies that PreambleDetected is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:       "fsk",
			irqMask:     IrqPreambleDetected,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "SyncWordValid_Single_FSK",
			desc:        "Verifies that SyncWordValid is globally enabled in the IRQ mask and correctly routed to the DIO1 pin, allowing the hardware to signal a valid FSK sync word detection.",
			modem:       "fsk",
			irqMask:     IrqSyncWordValid,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "HeaderValid_Single_FSK",
			desc:        "Verifies that attempting to enable HeaderValid while in FSK mode correctly returns an error, as it is a LoRa-specific interrupt.",
			modem:       "fsk",
			irqMask:     IrqHeaderValid,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "HeaderErr_Single_FSK",
			desc:        "Verifies that attempting to enable HeaderErr while in FSK mode correctly returns an error, as it is a LoRa-specific interrupt.",
			modem:       "fsk",
			irqMask:     IrqHeaderErr,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "CrcErr_Single_FSK",
			desc:        "Verifies that CrcErr is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:       "fsk",
			irqMask:     IrqCrcErr,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x40, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "CadDone_Single_FSK",
			desc:        "Verifies that attempting to use CAD interrupts while in FSK mode correctly returns an error.",
			modem:       "fsk",
			irqMask:     IrqCadDone,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "CadDetected_Single_FSK",
			desc:        "Verifies that attempting to use CAD interrupts while in FSK mode correctly returns an error.",
			modem:       "fsk",
			irqMask:     IrqCadDetected,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "Timeout_Single_FSK",
			desc:        "Verifies that Timeout is globally enabled in the IRQ mask and correctly routed to the DIO1 pin for FSK modulation.",
			modem:       "fsk",
			irqMask:     IrqTimeout,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "All_Single_FSK",
			desc:        "Verifies that enabling all interrupt bits (including reserved ones) is correctly rejected by the driver safety logic for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqAll,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "None_Single_FSK",
			desc:        "Verifies that a zero mask correctly disables all global interrupts and DIO routing for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqNone,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_Single_FSK",
			desc:        "Verifies that a standard set of FSK interrupts (TX, RX, Timeout, and CRC) is globally enabled and correctly routed to the DIO1 pin.",
			modem:       "fsk",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			irqMasks:    nil,
			txBytes:     []uint8{0x08, 0x02, 0x43, 0x02, 0x43, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},

		// LoRa IRQs - mixed
		{
			name:        "Mask_DIO1_LoRa",
			desc:        "Verifies that a custom IRQ mask is globally enabled and correctly routed exclusively to the DIO1 pin.",
			modem:       "lora",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678},
			txBytes:     []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Mask_DIO1_DIO2_LoRa",
			desc:        "Verifies that a global IRQ mask is globally enabled and correctly routed across both DIO1 and DIO2 pins simultaneously.",
			modem:       "lora",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678, 0x4321},
			txBytes:     []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x43, 0x21, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Mask_DIO1_DIO2_DIO3_LoRa",
			desc:        "Verifies that a global IRQ mask is globally enabled and correctly routed across all three pins (DIO1, DIO2, and DIO3) simultaneously.",
			modem:       "lora",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678, 0x4321, 0x8765},
			txBytes:     []uint8{0x08, 0x12, 0x34, 0x56, 0x78, 0x43, 0x21, 0x87, 0x65},
			expectError: false,
		},
		{
			name:        "DIO3_Only_LoRa",
			desc:        "Verifies that an interrupt can be routed exclusively to the DIO3 pin while DIO1 and DIO2 remain disabled, ensuring correct positional packing in the SPI payload.",
			modem:       "lora",
			irqMask:     IrqTxDone,
			irqMasks:    []IrqMask{0x0000, 0x0000, IrqTxDone},
			txBytes:     []uint8{0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "Masks_Overflow_LoRa",
			desc:        "Verifies that providing more than three DIO masks returns an error, preventing an invalid SPI transaction length.",
			modem:       "lora",
			irqMask:     IrqTxDone,
			irqMasks:    []IrqMask{IrqTxDone, IrqTxDone, IrqTxDone, IrqRxDone},
			txBytes:     nil,
			expectError: true,
		},

		// FSK IRQs - mixed
		{
			name:        "Mask_DIO1_FSK",
			desc:        "Verifies that a custom IRQ mask is globally enabled and correctly routed exclusively to the DIO1 pin for FSK mode.",
			modem:       "fsk",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678},
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "Mask_DIO1_DIO2_FSK",
			desc:        "Verifies that a global IRQ mask is globally enabled and correctly routed across both DIO1 and DIO2 pins simultaneously for FSK mode.",
			modem:       "fsk",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678, 0x4321},
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "Mask_DIO1_DIO2_DIO3_FSK",
			desc:        "Verifies that a global IRQ mask is globally enabled and correctly routed across all three pins (DIO1, DIO2, and DIO3) simultaneously for FSK mode.",
			modem:       "fsk",
			irqMask:     0x1234,
			irqMasks:    []IrqMask{0x5678, 0x4321, 0x8765},
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "DIO3_Only_FSK",
			desc:        "Verifies that an interrupt can be routed exclusively to the DIO3 pin while DIO1 and DIO2 remain disabled, ensuring correct positional packing in the SPI payload for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTxDone,
			irqMasks:    []IrqMask{0x0000, 0x0000, IrqTxDone},
			txBytes:     []uint8{0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "Masks_Overflow_FSK",
			desc:        "Verifies that providing more than three DIO masks returns an error, preventing an invalid SPI transaction length for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTxDone,
			irqMasks:    []IrqMask{IrqTxDone, IrqTxDone, IrqTxDone, IrqRxDone},
			txBytes:     nil,
			expectError: true,
		},

		// Misc
		{
			name:        "Error_UnknownModem",
			desc:        "Verifies that providing an invalid modem type string returns an error, as IRQ validation logic cannot be reliably applied.",
			modem:       "generic",
			irqMask:     IrqTxDone,
			irqMasks:    nil,
			txBytes:     nil,
			expectError: true,
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

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.3.3 GetIrqStatus
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
			commands:    []uint8{uint8(CmdGetIrqStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_LoRa",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the RxDone flag (bit 1) from the 16-bit interrupt status in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqRxDone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreambleDetected_LoRa",
			desc:        "Verifies that the driver correctly decodes the PreambleDetected flag (bit 2) from the MISO response stream while handling the byte offset correctly.",
			modem:       "lora",
			irqMask:     IrqPreambleDetected,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordValid_LoRa",
			desc:        "Verifies that the SyncWordValid flag (bit 3) is correctly decoded from the raw 16-bit IRQ status for the LoRa modem.",
			modem:       "lora",
			irqMask:     IrqSyncWordValid,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "HeaderValid_LoRa",
			desc:        "Verifies that the driver correctly decodes the HeaderValid flag (bit 4) from the 16-bit interrupt status, confirming a valid LoRa header reception.",
			modem:       "lora",
			irqMask:     IrqHeaderValid,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x10},
			expectError: false,
		},
		{
			name:        "HeaderErr_LoRa",
			desc:        "Verifies that the driver correctly decodes the HeaderErr flag (bit 5) to signal a corrupted LoRa header in the IRQ status.",
			modem:       "lora",
			irqMask:     IrqHeaderErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x20},
			expectError: false,
		},
		{
			name:        "CrcErr_LoRa",
			desc:        "Verifies that the driver correctly decodes the CrcErr flag (bit 6) from the interrupt status to signal a payload integrity failure.",
			modem:       "lora",
			irqMask:     IrqCrcErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_LoRa",
			desc:        "Verifies that the driver correctly decodes the CadDone flag (bit 7) marking the completion of a Channel Activity Detection operation.",
			modem:       "lora",
			irqMask:     IrqCadDone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x80},
			expectError: false,
		},
		{
			name:        "CadDetected_LoRa",
			desc:        "Verifies that the driver correctly decodes the CadDetected flag (bit 8), which resides in the upper byte of the 16-bit IRQ status response.",
			modem:       "lora",
			irqMask:     IrqCadDetected,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x01, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_LoRa",
			desc:        "Verifies that the driver correctly decodes the Timeout flag (bit 9) from the upper byte of the 16-bit IRQ status response.",
			modem:       "lora",
			irqMask:     IrqTimeout,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_LoRa",
			desc:        "Verifies that the driver correctly handles a response with no active interrupt flags, returning a clean zero status.",
			modem:       "lora",
			irqMask:     IrqNone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_LoRa",
			desc:        "Verifies that the driver can simultaneously decode multiple active interrupt flags (TX, RX, Timeout, CRC, Header) from a single 16-bit status response.",
			modem:       "lora",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
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
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_FSK",
			desc:        "Verifies that the driver correctly issues the GetIrqStatus opcode and decodes the RxDone flag (bit 1) from the 16-bit response specifically for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqRxDone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreambleDetected_FSK",
			desc:        "Verifies that the driver correctly decodes the PreambleDetected flag (bit 2) from the MISO response stream in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqPreambleDetected,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordValid_FSK",
			desc:        "Verifies that the driver correctly decodes the SyncWordValid flag (bit 3), which is a primary interrupt for FSK synchronization.",
			modem:       "fsk",
			irqMask:     IrqSyncWordValid,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "HeaderValid_FSK",
			desc:        "Verifies the decoding of bit 4 in FSK mode; ensures the driver handles the raw bitfield consistently across modem types.",
			modem:       "fsk",
			irqMask:     IrqHeaderValid,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x10},
			expectError: false,
		},
		{
			name:        "HeaderErr_FSK",
			desc:        "Verifies the decoding of bit 5 in FSK mode; ensures the driver handles the raw bitfield consistently across modem types.",
			modem:       "fsk",
			irqMask:     IrqHeaderErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x20},
			expectError: false,
		},
		{
			name:        "CrcErr_FSK",
			desc:        "Verifies that the driver correctly decodes the CrcErr flag (bit 6) to signal payload corruption in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqCrcErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_FSK",
			desc:        "Verifies the decoding of bit 7 in FSK mode; ensures consistent 16-bit status extraction.",
			modem:       "fsk",
			irqMask:     IrqCadDone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x80},
			expectError: false,
		},
		{
			name:        "CadDetected_FSK",
			desc:        "Verifies the decoding of bit 8 (upper byte) in FSK mode; ensures consistent 16-bit status extraction.",
			modem:       "fsk",
			irqMask:     IrqCadDetected,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x01, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_FSK",
			desc:        "Verifies that the driver correctly decodes the Timeout flag (bit 9) from the upper byte of the 16-bit IRQ status in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTimeout,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_FSK",
			desc:        "Verifies that the driver returns an empty status when the FSK modem reports no active interrupts.",
			modem:       "fsk",
			irqMask:     IrqNone,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          []uint8{0x00, 0x01, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_FSK",
			desc:        "Verifies the simultaneous decoding of a typical set of active FSK interrupt flags (TX, RX, Timeout, CRC) from the 16-bit response.",
			modem:       "fsk",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
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
			commands:    []uint8{uint8(CmdGetBufferStatus), OpCodeNop, OpCodeNop, OpCodeNop},
			tx:          []uint8{0x12, 0x00, 0x00, 0x00},
			rx:          nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.tx, spi.TxData)
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

// 13.3.4 ClearIrqStatus
func TestClearIrqStatus(t *testing.T) {
	tests := []struct {
		name        string
		desc        string
		modem       string
		irqMask     IrqMask
		txBytes     []uint8
		expectError bool
	}{
		// LoRa
		{
			name:        "TxDone_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the TxDone interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqTxDone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the RxDone interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqRxDone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreableDetected_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the PreambleDetected interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqPreambleDetected,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordVAlid_LoRa",
			desc:        "Verifies that the driver returns an error when attempting to clear the FSK-specific SyncWordValid flag while configured for LoRa mode.",
			modem:       "lora",
			irqMask:     IrqSyncWordValid,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "HeaderValid_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the HeaderValid interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqHeaderValid,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x10},
			expectError: false,
		},
		{
			name:        "HeaderErr_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the HeaderErr interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqHeaderErr,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x20},
			expectError: false,
		},
		{
			name:        "CrcErr_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the CrcErr interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqCrcErr,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the CadDone interrupt flag in LoRa mode.",
			modem:       "lora",
			irqMask:     IrqCadDone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x80},
			expectError: false,
		},
		{
			name:        "CadDetected_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the CadDetected interrupt flag, which resides in the upper byte of the payload.",
			modem:       "lora",
			irqMask:     IrqCadDetected,
			txBytes:     []uint8{0x02, 0x00, 0x01, 0x00},
			expectError: false,
		},
		{
			name:        "Timeout_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the Timeout interrupt flag, handling the upper byte of the payload correctly.",
			modem:       "lora",
			irqMask:     IrqTimeout,
			txBytes:     []uint8{0x02, 0x00, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_LoRa",
			desc:        "Verifies that sending an empty mask correctly results in a zeroed payload, effectively clearing no interrupts.",
			modem:       "lora",
			irqMask:     IrqNone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_LoRa",
			desc:        "Verifies that the driver correctly formats the SPI command to clear a standard combination of typical LoRa interrupts simultaneously.",
			modem:       "lora",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr | IrqHeaderErr,
			txBytes:     []uint8{0x02, 0x00, 0x02, 0x63},
			expectError: false,
		},
		{
			name:        "All_Single_LoRa",
			desc:        "Verifies that the driver's safety logic correctly rejects an attempt to clear all bits simultaneously, as it includes restricted or invalid flags.",
			modem:       "lora",
			irqMask:     IrqAll,
			txBytes:     nil,
			expectError: true,
		},

		// FSK
		{
			name:        "TxDone_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the TxDone interrupt flag in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTxDone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x01},
			expectError: false,
		},
		{
			name:        "RxDone_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the RxDone interrupt flag in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqRxDone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x02},
			expectError: false,
		},
		{
			name:        "PreableDetected_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the PreambleDetected interrupt flag in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqPreambleDetected,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x04},
			expectError: false,
		},
		{
			name:        "SyncWordVAlid_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the FSK-specific SyncWordValid interrupt flag.",
			modem:       "fsk",
			irqMask:     IrqSyncWordValid,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "HeaderValid_FSK",
			desc:        "Verifies that the driver returns an error when attempting to clear the LoRa-specific HeaderValid flag while configured for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqHeaderValid,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "HeaderErr_FSK",
			desc:        "Verifies that the driver returns an error when attempting to clear the LoRa-specific HeaderErr flag while configured for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqHeaderErr,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "CrcErr_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the CrcErr interrupt flag in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqCrcErr,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x40},
			expectError: false,
		},
		{
			name:        "CadDone_FSK",
			desc:        "Verifies that the driver returns an error when attempting to clear the LoRa-specific CadDone flag while configured for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqCadDone,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "CadDetected_FSK",
			desc:        "Verifies that the driver returns an error when attempting to clear the LoRa-specific CadDetected flag while configured for FSK mode.",
			modem:       "fsk",
			irqMask:     IrqCadDetected,
			txBytes:     nil,
			expectError: true,
		},
		{
			name:        "Timeout_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear the Timeout interrupt flag in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqTimeout,
			txBytes:     []uint8{0x02, 0x00, 0x02, 0x00},
			expectError: false,
		},
		{
			name:        "None_FSK",
			desc:        "Verifies that sending an empty mask correctly results in a zeroed payload, effectively clearing no interrupts in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqNone,
			txBytes:     []uint8{0x02, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Standard_Sane_FSK",
			desc:        "Verifies that the driver correctly formats the SPI command to clear a standard combination of typical FSK interrupts simultaneously.",
			modem:       "fsk",
			irqMask:     IrqTxDone | IrqRxDone | IrqTimeout | IrqCrcErr,
			txBytes:     []uint8{0x02, 0x00, 0x02, 0x43},
			expectError: false,
		},
		{
			name:        "All_Single_FSK",
			desc:        "Verifies that the driver's safety logic correctly rejects an attempt to clear all bits simultaneously in FSK mode.",
			modem:       "fsk",
			irqMask:     IrqAll,
			txBytes:     nil,
			expectError: true,
		},

		// Misc
		{
			name:        "Error_UnknownModem",
			desc:        "Verifies that providing an invalid modem type string returns an error, as IRQ validation logic cannot be reliably applied.",
			modem:       "generic",
			irqMask:     IrqTxDone,
			txBytes:     nil,
			expectError: true,
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

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.3.5 SetDIO2AsRfSwitchCtrl
func TestDIO2AsRfSwitchCtrl(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		enable  bool
		txBytes []uint8
	}{
		{
			name:    "EnableDIO2AsRfSwitch",
			desc:    "Verifies that the driver correctly configures the DIO2 pin to automatically control the external RF switch during transmission and reception cycles.",
			enable:  true,
			txBytes: []uint8{0x9D, 0x01},
		},
		{
			name:    "DisableDIO2AsRfSwitch",
			desc:    "Verifies that the driver correctly disables the automated external RF switch control on the DIO2 pin.",
			enable:  false,
			txBytes: []uint8{0x9D, 0x00},
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

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}

// 13.3.6 SetDIO3AsTCXOCtrl
func TestSetDIO3AsTCXOCtrl(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		voltage TcxoVoltage
		timeout uint32
		txBytes []uint8
	}{
		{
			name:    "DIO3Output1_6_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 1.6V with no stabilization delay.",
			voltage: Dio3Output1_6,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output1_7_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 1.7V with no stabilization delay.",
			voltage: Dio3Output1_7,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output1_8_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 1.8V with no stabilization delay.",
			voltage: Dio3Output1_8,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x02, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output2_2_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 2.2V with no stabilization delay.",
			voltage: Dio3Output2_2,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x03, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output2_4_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 2.4V with no stabilization delay.",
			voltage: Dio3Output2_4,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x04, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output2_7_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 2.7V with no stabilization delay.",
			voltage: Dio3Output2_7,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x05, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output3_0_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 3.0V with no stabilization delay.",
			voltage: Dio3Output3_0,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x06, 0x00, 0x00, 0x00},
		},
		{
			name:    "DIO3Output3_3_NoTimeout",
			desc:    "Verifies that the driver correctly sets the DIO3 output voltage to 3.3V with no stabilization delay.",
			voltage: Dio3Output3_3,
			timeout: 0x0000,
			txBytes: []uint8{0x97, 0x07, 0x00, 0x00, 0x00},
		},
		{
			name:    "TimeoutShift",
			desc:    "Verifies that a multi-byte timeout value is correctly split and packed into the 3-byte SPI payload in Big-Endian order.",
			voltage: 0x00,
			timeout: 0x123456,
			txBytes: []uint8{0x97, 0x00, 0x12, 0x34, 0x56},
		},
		{
			name:    "TimeoutMax24bit",
			desc:    "Verifies that the maximum possible 24-bit timeout value is correctly handled and sent in the SPI transaction.",
			voltage: 0x00,
			timeout: 0x00FFFFFF,
			txBytes: []uint8{0x97, 0x00, 0xFF, 0xFF, 0xFF},
		},
		{
			name:    "TimeoutMinNonZero",
			desc:    "Verifies that the driver correctly processes and packs the minimum possible non-zero timeout value into the least significant byte of the payload, ensuring no data loss during bitwise operations.",
			voltage: 0x00,
			timeout: 0x00000001,
			txBytes: []uint8{0x97, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:    "TimeoutMSBOnly",
			desc:    "Verifies that the driver correctly shifts and packs a timeout value containing only the most significant bit of the twenty-four-bit range, ensuring the upper byte boundary is handled accurately.",
			voltage: 0x00,
			timeout: 0x00800000,
			txBytes: []uint8{0x97, 0x00, 0x80, 0x00, 0x00},
		},
		{
			name:    "TimeoutOverflow",
			desc:    "Verifies the driver's safety logic: any timeout value exceeding 24 bits must be capped at 0xFFFFFF to prevent corruption of the SPI frame.",
			voltage: 0x00,
			timeout: 0x12FFFFFF,
			txBytes: []uint8{0x97, 0x00, 0xFF, 0xFF, 0xFF},
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

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}
		})
	}
}
