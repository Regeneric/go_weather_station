package ds18b20

import "time"

type Status int

const (
	Init                 Status = iota // 1
	ReadInit                           // 2
	Idle                               // 3
	StartConversion                    // 4
	WaitingForConversion               // 5
	StartRead                          // 6
	ReadingScratchpad                  // 7
	ProcessData                        // 8
	DataReady                          // 9
	DataError                          // 10
	DataProcessed
)

type Command byte

const (
	ConvertT        Command = 0x44
	WriteScratchpad Command = 0x4E
	CopyScratchpad  Command = 0x48
	RecallEE        Command = 0xBE
	ReadPowerSupply Command = 0xB4
)

type Config struct {
	Address          uint64
	Data             []uint8
	Temperature      float32
	DataCount        uint8
	State            uint16
	Resolution       uint16
	ConvertStartTime time.Time
	Queue            chan Data
}

type Data struct {
	Temperature float32 `json:"temperature"`
	Address     uint64  `json:"address"`
	Status      uint32  `json:"status"`
	Timestamp   int64   `json:"timestamp"`
}
