package bme280

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/Regeneric/go_weather_station/internal/config"
	"github.com/Regeneric/go_weather_station/pkg/model/bme"
	"periph.io/x/conn/v3/i2c"
)

// Official Bosch implementation
func compensateTemperature(adcT int32, config *bme.Config) (int32, int32) {
	log := slog.With("func", "compensateTemperature", "params", "(int32, int32, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Temperature compensation", "device", fmt.Sprintf("0x%02X", config.Address))

	var var1, var2, T, tFine int32

	var1 = (((adcT >> 3) - (int32(config.Params.DigT1) << 1)) * (int32(config.Params.DigT2))) >> 11
	var2 = (((((adcT >> 4) - (int32(config.Params.DigT1))) * ((adcT >> 4) - (int32(config.Params.DigT1)))) >> 12) * (int32(config.Params.DigT3))) >> 14
	tFine = var1 + var2
	T = (tFine*5 + 128) >> 8

	return T, tFine
}

// Official Bosch implementation
func compensatePressure(adcP, tFine int32, config *bme.Config) uint32 {
	log := slog.With("func", "compensatePressure", "params", "(int32, int32, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Pressure compensation", "device", fmt.Sprintf("0x%02X", config.Address))

	var var1, var2, P int64

	var1 = (int64(tFine)) - 128000
	var2 = var1 * var1 * int64(config.Params.DigP6)
	var2 = var2 + ((var1 * int64(config.Params.DigP5)) << 17)
	var2 = var2 + ((int64(config.Params.DigP4)) << 35)
	var1 = ((var1 * var1 * int64(config.Params.DigP3)) >> 8) + ((var1 * int64(config.Params.DigP2)) << 12)
	var1 = (((int64(1)) << 47) + var1) * (int64(config.Params.DigP1)) >> 33

	if var1 == 0 {
		return 0
	}

	P = 1048576 - int64(adcP)
	P = (((P << 31) - var2) * 3125) / var1
	var1 = ((int64(config.Params.DigP9)) * (P >> 13) * (P >> 13)) >> 25
	var2 = ((int64(config.Params.DigP8)) * P) >> 19
	P = ((P + var1 + var2) >> 8) + ((int64(config.Params.DigP7)) << 4)

	return uint32(P) // Q24.8 format, `24674867` equals to `24674867/256 == 96386.2 Pa == 963.862 hPa`
}

// Official Bosch implementation
func compensateHumidity(adcH, tFine int32, config *bme.Config) uint32 {
	log := slog.With("func", "compensateHumidity", "params", "(int32, int32, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Humidity compensation", "device", fmt.Sprintf("0x%02X", config.Address))

	var vX1u32r int32
	vX1u32r = (tFine - (int32(76800)))

	// What the fuck is going on here???
	vX1u32r = (((((adcH << 14) - ((int32(config.Params.DigH4)) << 20) - ((int32(config.Params.DigH5)) * vX1u32r)) +
		(int32(16384))) >> 15) * (((((((vX1u32r*(int32(config.Params.DigH6)))>>10)*
		(((vX1u32r*(int32(config.Params.DigH3)))>>11)+(int32(32768))))>>10)+(int32(2097152)))*
		(int32(config.Params.DigH2)) + 8192) >> 14))

	vX1u32r = (vX1u32r - (((((vX1u32r >> 15) * (vX1u32r >> 15)) >> 7) * (int32(config.Params.DigH1))) >> 4))
	vX1u32r = max(0, vX1u32r)         // vX1u32r = (vX1u32r < 0 ? 0 : vX1u32r)
	vX1u32r = min(vX1u32r, 419430400) // vX1u32r = (vX1u32r > 419430400 ? 419430400 : vX1u32r)
	// vX1u32r = min(419430400, max(0, vX1u32r)) // clamp() in one step - to be tested

	return uint32((vX1u32r >> 12))
}

func writeCommands(i2c i2c.BusCloser, config *bme.Config, commands []uint8) error {
	log := slog.With("func", "writeCommands", "params", "(i2c.BusCloser, *bme.Config, []uint8)", "package", "sensors", "module", "bme280")
	log.Debug("Writing commands to I2C device", "device", fmt.Sprintf("0x%02X", config.Address), "commands", fmt.Sprintf("% X", commands))

	if err := i2c.Tx(config.Address, commands, nil); err != nil {
		return fmt.Errorf("Failed to write commands [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	return nil
}

func readCalibrationData(i2c i2c.BusCloser, config *bme.Config) error {
	log := slog.With("func", "readCalibrationData", "params", "(i2c.BusCloser, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Reading data from I2C device", "device", fmt.Sprintf("0x%02X", config.Address))

	var buffer [26]uint8
	var commands [1]uint8
	commands[0] = bme.RegCalib00

	log.Debug("Reading calibration data", "device", fmt.Sprintf("0x%02X", config.Address), "commands", fmt.Sprintf("% X", commands))
	if err := i2c.Tx(config.Address, commands[:], buffer[:]); err != nil {
		return fmt.Errorf("Failed to read calibration data from registry [% X], from address 0x%02X: %w", commands, config.Address, err)
	}

	config.Params.DigT1 = (uint16(buffer[1])<<8 | uint16(buffer[0]))
	config.Params.DigT2 = (int16(buffer[3])<<8 | int16(buffer[2]))
	config.Params.DigT3 = (int16(buffer[5])<<8 | int16(buffer[4]))

	config.Params.DigP1 = (uint16(buffer[7])<<8 | uint16(buffer[6]))
	config.Params.DigP2 = (int16(buffer[9])<<8 | int16(buffer[8]))
	config.Params.DigP3 = (int16(buffer[11])<<8 | int16(buffer[10]))
	config.Params.DigP4 = (int16(buffer[13])<<8 | int16(buffer[12]))
	config.Params.DigP5 = (int16(buffer[15])<<8 | int16(buffer[14]))
	config.Params.DigP6 = (int16(buffer[17])<<8 | int16(buffer[16]))
	config.Params.DigP7 = (int16(buffer[19])<<8 | int16(buffer[18]))
	config.Params.DigP8 = (int16(buffer[21])<<8 | int16(buffer[20]))
	config.Params.DigP9 = (int16(buffer[23])<<8 | int16(buffer[22]))

	config.Params.DigH1 = buffer[25]
	commands[0] = bme.RegCalib26

	log.Debug("Reading calibration data", "device", fmt.Sprintf("0x%02X", config.Address), "commands", fmt.Sprintf("% X", commands))
	if err := i2c.Tx(config.Address, commands[:], buffer[:]); err != nil {
		return fmt.Errorf("Failed to read calibration data from registry [% X], from address 0x%02X: %w", commands, config.Address, err)
	}

	config.Params.DigH2 = (int16(buffer[1])<<8 | int16(buffer[0]))
	config.Params.DigH3 = buffer[2]
	config.Params.DigH4 = (int16(buffer[3]) << 4) | (int16(buffer[4]) & 0x0F)
	config.Params.DigH5 = (int16(buffer[5]) << 4) | (int16(buffer[4]) >> 4)
	config.Params.DigH6 = int8(buffer[6])

	return nil
}

func Init(i2c i2c.BusCloser, config *bme.Config) error {
	slog.Info("Initializing BME280 sensor...")
	slog.Info("Calibrating BME280 sensor...")

	log := slog.With("func", "Init", "params", "(i2c.BusCloser, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Initilizing sensor", "device", fmt.Sprintf("0x%02X", config.Address))

	if err := readCalibrationData(i2c, config); err != nil {
		return err
	}

	// Humidity
	var commands [2]uint8
	commands[0] = bme.RegCtrlHum
	commands[1] = config.HumiditySampling

	if err := i2c.Tx(config.Address, commands[:], nil); err != nil {
		return fmt.Errorf("Failed to write data [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	// IIR Coefficient
	commands[0] = bme.RegConfig
	commands[1] = config.IIRCoefficient

	if err := i2c.Tx(config.Address, commands[:], nil); err != nil {
		return fmt.Errorf("Failed to write data [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	// Pressure and temperature
	commands[0] = bme.RegCtrlMeas
	commands[1] = config.TemperatureAndPressureMode

	if err := i2c.Tx(config.Address, commands[:], nil); err != nil {
		return fmt.Errorf("Failed to write data [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	return nil
}

func initRead(i2c i2c.BusCloser, config *bme.Config) error {
	log := slog.With("func", "initRead", "params", "(i2c.BusCloser, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Initilizing read sequence", "device", fmt.Sprintf("0x%02X", config.Address))

	if config.Status == bme.ReadInProgress {
		return fmt.Errorf("Read from address 0x%02X in progress", config.Address)
	}
	config.Status = bme.ReadInProgress

	var commands [2]uint8
	commands[0] = bme.RegCtrlHum
	commands[1] = config.HumiditySampling

	if err := i2c.Tx(config.Address, commands[:], nil); err != nil {
		return fmt.Errorf("Failed to write data [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	commands[0] = bme.RegCtrlMeas
	commands[1] = config.TemperatureAndPressureMode

	if err := i2c.Tx(config.Address, commands[:], nil); err != nil {
		return fmt.Errorf("Failed to write data [% X] to address 0x%02X: %w", commands, config.Address, err)
	}

	return nil
}

func read(i2c i2c.BusCloser, config *bme.Config) error {
	log := slog.With("func", "Read", "params", "(i2c.BusCloser, *bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Reading data from sensor", "device", fmt.Sprintf("0x%02X", config.Address))

	reg := []uint8{0xF7}
	if err := i2c.Tx(config.Address, reg, config.RawData[:]); err != nil {
		return fmt.Errorf("Failed to write data [% X] or read from address 0x%02X: %w", reg, config.Address, err)
	}

	config.Status = bme.ReadSuccess
	return nil
}

func DewPoint(data bme.Data) float32 {
	log := slog.With("func", "DewPoint", "params", "(*bme.Data)", "package", "sensors", "module", "bme280")
	log.Debug("Calculating dew point")

	humidity := float64(data.Humidity)
	temperature := float64(data.Temperature)

	gamma := math.Log(humidity/100.0) + (17.625*temperature)/(243.04+temperature)
	dewPoint := (243.04 * gamma) / (17.625 - gamma)

	return float32(dewPoint)
}

func AbsoluteHumidity(data bme.Data) float32 {
	log := slog.With("func", "AbsoluteHumidity", "params", "(*bme.Data)", "package", "sensors", "module", "bme280")
	log.Debug("Calculating absolute humidity")

	// Magnus-Tetens formula
	humidity := float64(data.Humidity)
	temperature := float64(data.Temperature)

	pSat := 611.21 * math.Exp((17.67*temperature)/(temperature+243.5))
	pVapor := pSat * (humidity / 100.0)

	// The constant 2.167 is derived from the molar mass of water and the universal gas constant
	tempKelvin := temperature + 273.15
	absoluteHumidity := (2.167 * pVapor) / tempKelvin

	return float32(absoluteHumidity)
}

func ProcessData(config *bme.Config) bme.Data {
	log := slog.With("func", "ProcessData", "params", "(*bme.Config)", "package", "sensors", "module", "bme280")
	log.Debug("Processing collected raw data")

	rawPressure := (int32(config.RawData[0]) << 12) | (int32(config.RawData[1]) << 4) | (int32(config.RawData[2]) >> 4)
	rawTemp := (int32(config.RawData[3]) << 12) | (int32(config.RawData[4]) << 4) | (int32(config.RawData[5]) >> 4)
	rawHumidity := (int32(config.RawData[6]) << 8) | int32(config.RawData[7])

	temperature, tFine := compensateTemperature(rawTemp, config)
	pressure := compensatePressure(rawPressure, tFine, config)
	humidity := compensateHumidity(rawHumidity, tFine, config)

	data := bme.Data{
		Temperature: float32(temperature) / 100.0,
		Pressure:    (float32(pressure) / 256.0) / 100.0,
		Humidity:    float32(humidity) / 1024.0,
	}

	// We pass the whole structure, because it's easier to manage it somwhere else in the code
	data.AbsoluteHumidity = AbsoluteHumidity(data)
	data.DewPoint = DewPoint(data)

	return data
}

func AverageData(data []*bme.Data) bme.Data {
	log := slog.With("func", "AverageData", "params", "([]*bme.Data)", "package", "sensors", "module", "bme280")
	log.Debug("Average values from all sensors")

	if len(data) <= 0 {
		log.Warn("Data slice is empty")
		return bme.Data{}
	}

	if len(data) == 1 && data[0] != nil {
		log.Warn("Could not average data from only one sensor")
		return *data[0]
	}

	var sumRH, temp, pressure float32
	var validCount int
	for _, v := range data {
		if v == nil {
			log.Warn("Could not read data from sensor")
			continue
		}

		sumRH += v.Humidity
		temp += v.Temperature
		pressure += v.Pressure
		validCount++
	}

	if validCount == 0 {
		log.Error("Failed to read data from all sensors!")
		return bme.Data{}
	}

	fCount := float32(validCount)
	avg := bme.Data{
		Humidity:    sumRH / fCount,
		Temperature: temp / fCount,
		Pressure:    pressure / fCount,
	}
	avg.AbsoluteHumidity = AbsoluteHumidity(avg)
	avg.DewPoint = DewPoint(avg)

	return avg
}

func Run(ctx context.Context, params *bme.Params, queue chan<- bme.Data) {
	log := slog.With("func", "Run", "params", "(context.Context, *bme.Params, chan<- bme.Data)", "package", "sensors", "module", "bme280")
	log.Debug("Prepare, initialize and run sensor")

	if config.BME280Enable == false {
		log.Warn("BME280 has been disabled in the config file!", "enable", config.BME280Enable)
		return
	}

	// Error loop - if we cannot initilize sensor, wait 10 seconds and retry
	for {
		if err := Init(params.I2C, params.Config); err == nil {
			log.Info("BME280 sensor initilized successfully")
			break
		} else {
			log.Error("BM280 sensor init failed, retrying in 10s...", "err", err)
		}

		select {
		case <-time.After(10 * time.Second):
			continue // 10 seconds has passed, we try again
		case <-ctx.Done():
			return // Program has been closed somwhere else, we're done
		}
	}

	time.Sleep(100 * time.Millisecond)

	ticker := time.NewTicker(1 * time.Second) // 1 Hz polling rate
	defer ticker.Stop()

	// Main sensor loop
	for {
		select {
		case <-ticker.C:
			if err := initRead(params.I2C, params.Config); err != nil {
				log.Warn("BME280 sensor init read failed", "bus", params.I2C, "device", params.Config.Address, "err", err)
				continue
			}

			time.Sleep(15 * time.Millisecond)

			if err := read(params.I2C, params.Config); err != nil {
				log.Warn("BME280 sensor read failed", "bus", params.I2C, "device", params.Config.Address, "err", err)
				continue
			}

			if params.Config.Status != bme.ReadSuccess {
				log.Warn("BME280 sensor read failed", "bus", params.I2C, "device", params.Config.Address, "status", params.Config.Status)
				continue
			}

			data := ProcessData(params.Config)

			select {
			case queue <- data:
				continue // Data has been sent to the queue
			default:
				log.Warn("BME 280 data queue full, dropping packet", "bus", params.I2C, "device", params.Config.Address)
			}
		case <-ctx.Done():
			return
		}
	}
}
