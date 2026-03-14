package sgp_manager

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/Regeneric/iot-drivers/libs/sgp30"
)

type SGP struct {
	HW   *sgp30.Device
	MU   sync.Mutex
	ECO2 uint16
	TVOC uint16
	Err  error
}

func (s *SGP) Run(ctx context.Context) error {
	log := slog.With("func", "SGP.Run()", "params", "(context.Context)", "return", "(-)", "package", "sgp_manager")
	log.Info("[ SGP ] Sensor event loop")

	if ctx == nil || reflect.ValueOf(ctx).IsNil() {
		return fmt.Errorf("[ SGP ] Sensor state improper; ctx is nil")
	}

	if s.HW == nil {
		return fmt.Errorf("[ SGP ] Sensor state improper; HW is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// 1 Hz for accurate readings
	measureTicker := time.NewTicker(1 * time.Second)
	defer measureTicker.Stop()

	// IAQ baseline value - save every hour, change once a week
	calibrationTicker := time.NewTicker(1 * time.Hour)
	defer calibrationTicker.Stop()

	baseline := make([]uint8, 6)
	buffer := make([]uint8, 6)

	// 1 Hz measure loop
	for range measureTicker.C {
		err := s.HW.MeasureIaq(buffer)
		if err != nil {
			continue
		}

		eco2 := uint16(buffer[0])<<8 | uint16(buffer[1])
		tvoc := uint16(buffer[3])<<8 | uint16(buffer[4])

		s.MU.Lock()
		s.ECO2 = eco2
		s.TVOC = tvoc
		s.MU.Unlock()
	}

	// Calibration loop
	for range calibrationTicker.C {
		if err := s.HW.GetIaqBaseline(baseline); err != nil {
			log.Error("[ SGP ] Could not read IAQ baseline value", "error", err)
		} else {
			filename := fmt.Sprintf("sgp30_baseline_%s.bin", s.HW.Config.Name)
			if err := os.WriteFile(filename, baseline, 0644); err != nil {
				log.Error("[ SGP ] Could not save IAQ baseline value to file", "error", err)
			}
		}
	}

	return nil
}
