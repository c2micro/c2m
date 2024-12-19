package utils

import (
	"math/rand"
	"strings"

	"github.com/c2micro/c2m/internal/constants"
)

// RandUint32 генерация рандомного числа
func RandUint32() uint32 {
	return rand.Uint32()
}

// RandString генерация рандомной строки заданной длины
func RandString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// rand.Int63() генерирует 63 рандомных бита - этого достаточно для letterIdxMax символов
	for i, cache, remain := n-1, rand.Int63(), constants.LetterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), constants.LetterIdxMax
		}
		if idx := int(cache & constants.LetterIdxMask); idx < len(constants.LetterBytes) {
			sb.WriteByte(constants.LetterBytes[idx])
			i--
		}
		cache >>= constants.LetterIdxBits
		remain--
	}

	return sb.String()
}
