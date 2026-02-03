package pms

type Status int

const (
	Init         Status = 1
	ReadInit     Status = 2
	PacketLength Status = 32
	StartByte1   Status = 0x42
	StartByte2   Status = 0x4D
)

type Config struct {
	RawData []uint8
	Status  uint8
	Queue   chan Data
}

type Data struct {
	PM1       uint16 `json:"pm1"`
	PM2_5     uint16 `json:"pm2_5"`
	PM10      uint16 `json:"pm10"`
	Timestamp int64  `json:"timestamp"`
}
