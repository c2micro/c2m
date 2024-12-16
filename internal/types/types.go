package types

import (
	"database/sql/driver"
	"fmt"
	"net"
)

// Inet кастомный тип для хранения одного IP адреса
type Inet struct {
	net.IP
}

// Scan имплементация Scanner интерфейса
func (i *Inet) Scan(value any) (err error) {
	switch v := value.(type) {
	case nil:
	case []byte:
		if i.IP = net.ParseIP(string(v)); i.IP == nil {
			err = fmt.Errorf("invalid value of ip %q", v)
		}
	case string:
		if i.IP = net.ParseIP(v); i.IP == nil {
			err = fmt.Errorf("invalid value of ip %q", v)
		}
	default:
		err = fmt.Errorf("unexpected type %T", v)
	}
	return
}

// Value имплементация Valuer интерфейса драйвера
func (i Inet) Value() (driver.Value, error) {
	return i.IP.String(), nil
}

// String приведение к типу строки
func (i Inet) String() string {
	if i.IP == nil {
		return ""
	}
	return i.IP.String()
}
