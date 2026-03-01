package sgp30

type MockI2C struct {
	TxData      []uint8
	RxData      []uint8
	ReturnError error
}

func (m *MockI2C) Tx(addr uint16, w, r []byte) error {
	if m.ReturnError != nil {
		return m.ReturnError
	}

	m.TxData = append(m.TxData, uint8(addr>>8), uint8(addr))
	m.TxData = append(m.TxData, w...)

	if r != nil && len(m.RxData) > 0 {
		copy(r, m.RxData)
	}
	return nil
}

func (m *MockI2C) String() string { return "MockSPI" }
