package sx1262

import (
	"sync"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
)

type LoraModem struct {
	hw        *lora.Device
	RxQueue   chan []uint8
	TxQueue   chan []uint8
	irqQueue  chan struct{}
	stopQueue chan struct{}
	mu        sync.Mutex
}

type RxBufferStatus struct {
	PayloadLengthRx      uint8
	RxStartBufferPointer uint8
}

type LoraModemStatus struct {
	CommandStatus      uint8
	ChipMode           uint8
	PacketsReceived    uint16
	PacketsCrcError    uint16
	PacketsHeaderError uint16
}
