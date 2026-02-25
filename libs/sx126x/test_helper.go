package sx126x

import (
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type MockSPI struct {
	TxData      []uint8
	RxData      []uint8
	ReturnError error
}

func (m *MockSPI) Tx(w, r []uint8) error {
	m.TxData = append(m.TxData, w...)
	if r != nil && len(m.RxData) > 0 {
		copy(r, m.RxData)
	}
	return nil
}

func (m *MockSPI) Duplex() conn.Duplex            { return conn.Half }
func (m *MockSPI) TxPackets(p []spi.Packet) error { return nil }
func (m *MockSPI) String() string                 { return "MockSPI" }
func (m *MockSPI) Baud() physic.Frequency         { return 0 }
