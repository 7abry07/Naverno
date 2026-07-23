package storage

type Storage interface {
	Write(off uint64, data []byte) error
	Read(off uint64, length uint32) ([]byte, error)
}
