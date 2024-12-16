package listener

import (
	"net"
)

// ConfigV1 структура конфигурации listener сервера
type ConfigV1 struct {
	IP   net.IP `json:"ip" validate:"required"`
	Port int    `json:"port" validate:"required,port"`
}
