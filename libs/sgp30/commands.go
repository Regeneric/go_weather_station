package sgp30

import (
	"fmt"
	"log/slog"
)

func (d *Device) IaqInit() error {
	log := slog.With("func", "Device.IaqInit()", "params", "(-)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 sensor init")

	if err := d.I2C.Tx(uint16(RegIaqInit), nil, nil); err != nil {
		return fmt.Errorf("Could not send init command to SGP30 sensor: %w", err)
	}

	log.Info("Init command send to sensor", "address", RegIaqInit)
	return nil
}

func (d *Device) MeasureIaq(data []uint8) error {
	log := slog.With("func", "Device.MeasureIaq()", "params", "()", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 measure eCO2 and TVOC values")

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := d.I2C.Tx(uint16(RegMeasureIaq), nil, data); err != nil {
		return fmt.Errorf("Could not send measure command to SGP30 sensor: %w", err)
	}

	// Function validates CRC after it was received to `data` buffer (duh)
	// So it is possible to ignore returned error and look at raw data
	if err := validateCRC(data[0:3]); err != nil {
		return fmt.Errorf("eCO2 CRC validation error: %w", err)
	}
	if err := validateCRC(data[3:6]); err != nil {
		return fmt.Errorf("TVOC CRC validation error: %w", err)
	}

	log.Info("Measure command send to sensor", "address", RegMeasureIaq)
	return nil
}

func (d *Device) GetIaqBaseline() error           { return nil }
func (d *Device) SetIaqBaseline() error           { return nil }
func (d *Device) SetAbsoluteHumidity() error      { return nil }
func (d *Device) MeasureTest() error              { return nil }
func (d *Device) GetFeatureSet() error            { return nil }
func (d *Device) MeasureRaw() error               { return nil }
func (d *Device) GetTvocInceptiveBaseline() error { return nil }
func (d *Device) SetTvocBaseline() error          { return nil }
func (d *Device) SoftReset() error                { return nil }
func (d *Device) GetSerialId() error              { return nil }

func calculateCRC(data []uint8) (uint8, error) {
	log := slog.With("func", "calculateCRC()", "params", "([]uint8)", "return", "(uint8, error)", "lib", "sgp30")
	log.Debug("SGP30 calculate CRC checksum")

	dataFrameLength := 2
	if len(data) == 0 || len(data) > dataFrameLength {
		return 0, fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	var crc uint8 = 0xFF  // 0b11111111
	var mask uint8 = 0x31 // 0b00110001

	for _, v := range data[0:2] { // <0,2)
		crc = crc ^ v // XOR
		for range 8 {
			msbit := crc & 0x80 // Most significant BIT
			crc = crc << 1
			if msbit != 0 {
				crc = crc ^ mask // XOR
			}
		}
	}

	return crc, nil
}

func validateCRC(data []uint8) error {
	log := slog.With("func", "validateCRC()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 validate CRC checksum")

	dataFrameLength := 3
	if len(data) == 0 || len(data) > dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	crc, err := calculateCRC(data[0:2])
	if err != nil {
		return err
	}

	if crc != data[2] {
		return fmt.Errorf("Checksum invalid.\nExpected: [%# x ]\nGot:      [%# x ]", data[2], crc)
	}

	return nil
}
