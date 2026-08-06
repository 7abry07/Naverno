package discardstorage

type DiscardStorage struct{}

func New() *DiscardStorage {
	return &DiscardStorage{}
}

func (s *DiscardStorage) Write(off uint64, data []byte) error {
	return nil
}
func (s *DiscardStorage) Read(off uint64, length uint64) ([]byte, error) {
	return make([]byte, length), nil
}
