package system

type SensorPacket struct {
	SourceID  string      `json:"deviceId"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

const (
	PacketTypeBME = 0x00
)

type PayloadBME280 struct {
}
