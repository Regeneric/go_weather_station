package dht

type Status int

const (
	Init           Status = iota // 0
	ReadInit                     // 1
	ReadSuccess                  // 2
	ReadInProgress               // 3
	BadChecksum                  // 4
)

type Model int

const (
	DHT11 Model = iota // 0
	DHT20              // 1
	DHT22              // 2
)

type Config struct {
	GPIO    uint8
	Address uint8
	Data    []uint8
	RawData []uint8
	Status  uint8
	Type    uint8
	Queue   chan Data
}

type Data struct {
	Sensor           string  `json:"sensor"`
	Pressure         float32 `json:"pressure"`
	Temperature      float32 `json:"temperature"`
	DewPoint         float32 `json:"dewPoint"`
	Humidity         float32 `json:"humidity"`
	AbsoluteHumidity float32 `json:"absoluteHumidity"`
	Timestamp        int64   `json:"timestamp"`
}
