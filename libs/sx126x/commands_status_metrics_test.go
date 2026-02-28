package sx126x

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

// GetPacketStatus() (PacketStatus, error)
// GetRssiInst() (int8, error)

func init() {
	discardHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(discardHandler))
}

// 13.5.1 GetStatus
func TestGetStatus(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		status  ModemStatus
		tx      []uint8
		rx      []uint8
		txBytes []uint8
	}{
		// Command
		{
			name: "Command_DataAvailable",
			desc: "Verifies decoding of the Data Available command status, indicating successful packet reception.",
			status: ModemStatus{
				Command: StatusDataAvailable,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x04},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Command_Timeout",
			desc: "Verifies decoding of the Command Timeout status, indicating the SPI watchdog expired.",
			status: ModemStatus{
				Command: StatusCmdTimeout,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x06},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Command_ProcessingError",
			desc: "Verifies decoding of the Command Processing Error status, indicating an invalid opcode or parameters.",
			status: ModemStatus{
				Command: StatusCmdProcessingError,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x08},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Command_ExecuteError",
			desc: "Verifies decoding of the Command Execute Error status, indicating the chip cannot perform the requested action.",
			status: ModemStatus{
				Command: StatusCmdExecuteError,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x0A},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Command_TxDone",
			desc: "Verifies decoding of the TX Done command status, indicating successful transmission completion.",
			status: ModemStatus{
				Command: StatusCmdTxDone,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x0C},
			txBytes: []uint8{0xC0, 0x00},
		},

		// Chip
		{
			name: "Chip_StandbyRc",
			desc: "Verifies decoding of the Standby RC chip mode.",
			status: ModemStatus{
				ChipMode: StatusModeStdbyRc,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x20},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Chip_StandbyXosc",
			desc: "Verifies decoding of the Standby XOSC chip mode.",
			status: ModemStatus{
				ChipMode: StatusModeStdbyXosc,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x30},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Chip_FS",
			desc: "Verifies decoding of the Frequency Synthesis chip mode.",
			status: ModemStatus{
				ChipMode: StatusModeFs,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x40},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Chip_RX",
			desc: "Verifies decoding of the Receive chip mode.",
			status: ModemStatus{
				ChipMode: StatusModeRx,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x50},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Chip_TX",
			desc: "Verifies decoding of the Transmit chip mode.",
			status: ModemStatus{
				ChipMode: StatusModeTx,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x60},
			txBytes: []uint8{0xC0, 0x00},
		},

		// Combined
		{
			name: "Combined_RX_DataAvailable",
			desc: "Verifies that the driver correctly extracts both Chip Mode and Command Status when they are combined in a single status byte.",
			status: ModemStatus{
				Command:  StatusDataAvailable,
				ChipMode: StatusModeRx,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x54},
			txBytes: []uint8{0xC0, 0x00},
		},
		{
			name: "Combined_TX_TxDone",
			desc: "Verifies decoding of a completed transmission state where the chip signals TX mode alongside a TX Done command status.",
			status: ModemStatus{
				Command:  StatusCmdTxDone,
				ChipMode: StatusModeTx,
			},
			tx:      []uint8{uint8(CmdGetStatus), OpCodeNop},
			rx:      []uint8{0x00, 0x6C},
			txBytes: []uint8{0xC0, 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			dev := Device{SPI: &spi}

			status, err := dev.GetStatus()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetStatus returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}

			if status != tc.status {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: %v\nGot:      %v", tc.desc, tc.status, status)
			}
		})
	}
}

// 13.5.3 GetPacketStatus
func TestGetPacketStatus(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		status  PacketStatus
		tx      []uint8
		rx      []uint8
		txBytes []uint8
	}{
		// LoRa
		{
			name: "RssiPkt100_SnrPkt32_SignalRssi100_LoRa",
			desc: "",
			status: PacketStatus{
				SignalStrength:         -50,
				SnRRatio:               8,
				DenoisedSignalStrength: -50,
			},
			tx:      []uint8{uint8(CmdGetPacketStatus), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x64, 0x20, 0x64},
			txBytes: []uint8{0x14, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "RssiPkt0_SnrPkt0_SignalRssi0_LoRa",
			desc: "",
			status: PacketStatus{
				SignalStrength:         0,
				SnRRatio:               0,
				DenoisedSignalStrength: 0,
			},
			tx:      []uint8{uint8(CmdGetPacketStatus), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x00, 0x00, 0x00},
			txBytes: []uint8{0x14, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "RssiPktMax8bit_SnrPktMax8bit_SignalRssiMax8bit_LoRa",
			desc: "",
			status: PacketStatus{
				SignalStrength:         -127,
				SnRRatio:               -0.25,
				DenoisedSignalStrength: -127,
			},
			tx:      []uint8{uint8(CmdGetPacketStatus), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0xFF, 0xFF, 0xFF},
			txBytes: []uint8{0x14, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "RssiPktShift_SnrPktShift_SignalRssiShift_LoRa",
			desc: "",
			status: PacketStatus{
				SignalStrength:         -9,
				SnRRatio:               13,
				DenoisedSignalStrength: -43,
			},
			tx:      []uint8{uint8(CmdGetPacketStatus), OpCodeNop, OpCodeNop, OpCodeNop, OpCodeNop},
			rx:      []uint8{0x00, 0x01, 0x12, 0x34, 0x56},
			txBytes: []uint8{0x14, 0x00, 0x00, 0x00, 0x00},
		},

		// FSK - TODO
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spi := MockSPI{RxData: tc.rx}
			dev := Device{SPI: &spi}

			status, err := dev.GetPacketStatus()

			if err != nil {
				t.Fatalf("FAIL: %s\nGetPacketStatus returned: %v", tc.desc, err)
			}

			if !bytes.Equal(spi.TxData, tc.txBytes) {
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.txBytes, spi.TxData)
			}

			if status != tc.status {
				t.Errorf("FAIL: %s\nWrong bytes received from SPI!\nExpected: %v\nGot:      %v", tc.desc, tc.status, status)
			}
		})
	}
}
