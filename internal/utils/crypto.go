package utils

import (
	"crypto/elliptic"
)

// ChooseEC выбор эллиптической кривой в зависимости от длины
func ChooseEC(l int) elliptic.Curve {
	switch l {
	case 224:
		return elliptic.P224()
	case 256:
		return elliptic.P256()
	case 384:
		return elliptic.P384()
	case 521:
		return elliptic.P521()
	default:
		return nil
	}
}
