package utils

import (
	"encoding/binary"

	"github.com/c2micro/c2msrv/internal/constants"
	"github.com/orisano/wyhash"
)

// CalcHash вычисление не криптографического хэша
func CalcHash(data []byte) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, wyhash.Sum64(constants.WyhashSeed, data))
	return b
}
