package utils

import (
	"os"
)

// EnvOr если есть значение энв. переменной по ключу - перезаписывает def
func EnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
