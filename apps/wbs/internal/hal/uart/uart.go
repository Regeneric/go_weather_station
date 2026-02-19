package uart

import (
	"fmt"
	"log/slog"
	"wbs/internal/config"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/uart"
	"periph.io/x/conn/v3/uart/uartreg"
	"periph.io/x/host/v3"
)

func Init(device string) (uart.PortCloser, error) {
	log := slog.With("func", "Init()", "params", "(string)", "return", "(uart.PortCloser, error)", "package", "uart")
	log.Info("Initializing UART bus", "device", device)

	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("[UART] Host init failed: %w", err)
	}

	bus, err := uartreg.Open(device)
	if err != nil {
		return nil, fmt.Errorf("[UART] Failed to open UART bus %s: %w", device, err)
	}

	return bus, nil
}

func Setup(cfg *config.UART) (map[string]conn.Conn, func(), error) {
	log := slog.With("func", "Setup()", "params", "(*config.UART)", "return", "(map[string]conn.Conn, func(), error)", "package", "uart")
	log.Info("UART bus setup")

	if cfg.Enable == false {
		return nil, func() {}, fmt.Errorf("UART bus disabled in the config")
	}

	conns := make(map[string]conn.Conn)
	var closers []func() error

	cleanup := func() {
		for i, c := range closers {
			slog.Debug("Closing UART bus connection...", "connection", i)
			_ = c()
		}
	}

	for key, dev := range cfg.Devices {
		if dev.Enable == false {
			continue
		}

		port, err := Init(dev.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to init UART %s (%s): %w", key, dev.Name, err)
		}
		closers = append(closers, port.Close)

		byteToFlow := map[uint8]uart.Flow{
			0: uart.NoFlow,
			1: uart.XOnXOff,
			2: uart.RTSCTS,
		}

		flow, ok := byteToFlow[dev.DataFlow]
		if !ok {
			flow = uart.NoFlow
			log.Warn("Unknown flow option", "flow", dev.DataFlow)
			log.Warn("Limiting FLOW to NoFlow")
		}

		byteToParity := map[string]uart.Parity{
			"N": uart.NoParity,
			"O": uart.Odd,
			"E": uart.Even,
			"M": uart.Mark,
			"S": uart.Space,
		}

		parity, ok := byteToParity[dev.Parity]
		if !ok {
			parity = uart.NoParity
			log.Warn("Unknown parity option", "parity", dev.Parity)
			log.Warn("Limiting PARITY to NoParity")
		}

		conn, err := port.Connect(physic.Frequency(dev.Speed), dev.StopBit, parity, flow, dev.DataLength)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("Failed to configure UART %s (%s): %w", key, dev.Name, err)
		}

		conns[key] = conn
		slog.Debug("UART device configured", "key", key, "name", dev.Name, "speed", dev.Speed, "stopBit", dev.StopBit, "parity", parity, "dataFlow", flow, "dataLength", dev.DataLength)
	}

	return conns, cleanup, nil
}
