package sx126x

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
