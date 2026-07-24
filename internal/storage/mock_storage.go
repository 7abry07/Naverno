package storage

type MockStorage struct{}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (s *MockStorage) Write(off uint64, data []byte) error {
	return nil
}
func (s *MockStorage) Read(off uint64, length uint32) ([]byte, error) {
	return make([]byte, length), nil
}
