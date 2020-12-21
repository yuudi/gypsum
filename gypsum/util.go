package gypsum

import (
	"encoding/binary"
)

func ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func ToUint(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
