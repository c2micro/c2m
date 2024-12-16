package operator

import (
	"net"
)

// ConfigV1 структура конфигурации operator сервера
type ConfigV1 struct {
	IP   net.IP `json:"ip" validate:"required"`
	Port int    `json:"port" validate:"required,port"`
}
