package uart

import (
	"fmt"
	"log/slog"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/uart"
	"periph.io/x/conn/v3/uart/uartreg"
	"periph.io/x/host/v3"
)

func Init(device string, baudRate int) (uart.PortCloser, error) {
	log := slog.With("func", "Init", "params", "(string, int)", "package", "bus", "module", "uart")
	log.Info("Initializing UART interface", "device", device, "baud", baudRate)

	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[UART] Host init failed: %w", err)
	}

	if _, err := driverreg.Init(); err != nil {
		return nil, fmt.Errorf("[UART] Driver init failed: %w", err)
	}

	log.Debug("[UART] Host init successful")

	port, err := uartreg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[UART] Failed to open UART device %s: %w", device, err)
	}

	speed := physic.Frequency(baudRate) * physic.Hertz
	if _, err := port.Connect(speed, uart.One, uart.NoParity, uart.NoFlow, 8); err != nil {
		port.Close()
		return nil, fmt.Errorf("[UART] Failed to open UART device: %w", err)
	}

	return port, nil
}

func Scan() error {
	slog.Info("--- SCANING AVAILABLE UART PORTS ---")
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("[UART] Host init failed: %w", err)
	}

	if _, err := driverreg.Init(); err != nil {
		return fmt.Errorf("[UART] Driver init failed: %w", err)
	}

	ports := uartreg.All()
	if len(ports) == 0 {
		slog.Warn("No UART ports found")
	} else {
		for _, port := range ports {
			slog.Info("Found UART port", "name", port.Name, "number", port.Number, "aliases", port.Aliases, "open", port.Open)
		}
	}
	slog.Info("-------------------------------------")
	return nil
}
