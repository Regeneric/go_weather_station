package sx1262

import (
	"sync"

	"github.com/Regeneric/go_weather_station/pkg/model/lora"
)

type LoraModem struct {
	hw      *lora.Device
	RxQueue chan []uint8
	TxQueue chan []uint8
	irqChan chan []uint8
	mu      sync.Mutex
}

type RxBufferStatus struct {
	PayloadLengthRx      uint8
	RxStartBufferPointer uint8
}
