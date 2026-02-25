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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
		t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", desc, expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
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
				t.Errorf("FAIL: %s\nWrong bytes send to SPI!\nExpected: [%# x]\nSent:     [%# x]", tc.desc, tc.expectedBytes, spi.TxData)
			}
		})
	}
}
