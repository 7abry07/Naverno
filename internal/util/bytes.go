package util

func Align(n, alignment uint64) uint64 {
	return (n + alignment - 1) / alignment * alignment
}
