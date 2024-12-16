package management

import (
	"net"
)

// ConfigV1 структура конфигурации management сервера
type ConfigV1 struct {
	IP   net.IP `json:"ip" validate:"required"`
	Port int    `json:"port" validate:"required,port"`
}
