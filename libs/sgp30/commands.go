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
	log := slog.With("func", "Device.MeasureIaq()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 measure eCO2 and TVOC values")

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := d.I2C.Tx(uint16(RegMeasureIaq), nil, data); err != nil {
		return fmt.Errorf("Could not send measure command to SGP30 sensor: %w", err)
	}

	// Function validates CRC after it was received to `data` buffer (duh)
	// So it is possible to ignore returned error and look directly at raw data
	if err := validateCRC(data[0:3]); err != nil {
		return fmt.Errorf("eCO2 CRC validation error: %w", err)
	}
	if err := validateCRC(data[3:6]); err != nil {
		return fmt.Errorf("TVOC CRC validation error: %w", err)
	}

	log.Info("Measure command send to sensor", "address", RegMeasureIaq)
	return nil
}

// 6 bytes: (eCO2_MSB, eCO2_LSB, eCO2_CRC, TVOC_MSB, TVOC_LSB, TVOC_CRC)
func (d *Device) GetIaqBaseline(data []uint8) error {
	log := slog.With("func", "Device.GetIaqBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 get IAQ baseline calibration value")

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := d.I2C.Tx(uint16(RegGetIaqBaseline), nil, data); err != nil {
		return fmt.Errorf("Could not get IAQ baseline values from SGP30 sensor: %w", err)
	}

	// Function validates CRC after it was received to `data` buffer (duh)
	// So it is possible to ignore returned error and look directly at raw data
	if err := validateCRC(data[0:3]); err != nil {
		return fmt.Errorf("eCO2 CRC validation error: %w", err)
	}
	if err := validateCRC(data[3:6]); err != nil {
		return fmt.Errorf("TVOC CRC validation error: %w", err)
	}

	log.Info("Get IAQ baseline command send to sensor", "address", RegGetIaqBaseline)
	return nil
}

// 6 bytes in `data`	: (eCO2_MSB, eCO2_LSB, eCO2_CRC, TVOC_MSB, TVOC_LSB, TVOC_CRC)
// 6 bytes sent to `i2c`: (TVOC_MSB, TVOC_LSB, TVOC_CRC, eCO2_MSB, eCO2_LSB, eCO2_CRC)
func (d *Device) SetIaqBaseline(data []uint8) error {
	log := slog.With("func", "Device.SetIaqBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 set IAQ baseline calibration value")

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := validateCRC(data[0:3]); err != nil {
		return fmt.Errorf("eCO2 CRC validation error: %w", err)
	}
	if err := validateCRC(data[3:6]); err != nil {
		return fmt.Errorf("TVOC CRC validation error: %w", err)
	}

	// TVOC_MSB, TVOC_LSB, TVOC_CRC, eCO2_MSB, eCO2_LSB, eCO2_CRC)
	baseline := []uint8{data[3], data[4], data[5], data[0], data[1], data[2]}
	if err := d.I2C.Tx(uint16(RegSetIaqBaseline), baseline, nil); err != nil {
		return fmt.Errorf("Could not set IAQ baseline values to SGP30 sensor: %w", err)
	}

	log.Info("IAQ set baseline command send to sensor", "address", RegSetIaqBaseline)
	return nil
}

func (d *Device) SetAbsoluteHumidity(data []uint8) error {
	log := slog.With("func", "Device.SetAbsoluteHumidity()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 set absolute humidity calibration value")

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := validateCRC(data); err != nil {
		return fmt.Errorf("CRC validation error: %w", err)
	}

	if err := d.I2C.Tx(uint16(RegSetAbsoluteHumidity), data, nil); err != nil {
		return fmt.Errorf("Could not set absolute humidity calibration values to SGP30 sensor: %w", err)
	}

	log.Info("Absolute humidity calibration command send to sensor", "address", RegSetAbsoluteHumidity)
	return nil
}

// Page 10 Measure Test
func (d *Device) MeasureTest(data []uint8) error {
	log := slog.With("func", "Device.MeasureTest()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("SGP30 sensor self-test")

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	if err := d.I2C.Tx(uint16(RegMeasureTest), nil, data); err != nil {
		return fmt.Errorf("Could not send measure test command to SGP30 sensor: %w", err)
	}

	testValue := uint16(data[0])<<8 | uint16(data[1])
	if testValue != uint16(MeasureTest) {
		return fmt.Errorf("Measure test returned unexpected value [%# x]", testValue)
	}

	if err := validateCRC(data); err != nil {
		return fmt.Errorf("CRC validation error: %w", err)
	}

	log.Info("Measure test command send to sensor", "address", RegMeasureTest)
	return nil
}

func (d *Device) GetFeatureSet() error            { return nil }
func (d *Device) MeasureRaw() error               { return nil }
func (d *Device) GetTvocInceptiveBaseline() error { return nil }
func (d *Device) SetTvocBaseline() error          { return nil }
func (d *Device) SoftReset() error                { return nil }
func (d *Device) GetSerialId() error              { return nil }

// 6.6 Checksum Calculation
func calculateCRC(data []uint8) (uint8, error) {
	log := slog.With("func", "calculateCRC()", "params", "([]uint8)", "return", "(uint8, error)", "lib", "sgp30")
	log.Debug("SGP30 calculate CRC checksum")

	dataFrameLength := 2
	if len(data) == 0 || len(data) > dataFrameLength {
		return 0, fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	var crc uint8 = uint8(CrcBase)  // 0b11111111
	var mask uint8 = uint8(CrcMask) // 0b00110001

	for _, v := range data[0:2] { // <0,2)
		crc = crc ^ v // XOR
		for range 8 {
			msbit := crc & uint8(CrcMsbit) // Most significant BIT
			crc = crc << 1
			if msbit != 0 {
				crc = crc ^ mask // XOR
			}
		}
	}

	return crc, nil
}

// 6.6 Checksum Calculation - TODO: dynamic validation, so it detects if we sent 3, 6, 9 or more bytes
// Dataframe: (DATA0, DATA1, CRC)
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
