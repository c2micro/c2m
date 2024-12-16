package defaults

import (
	"database/sql/driver"
)

type TaskMessage uint8

const (
	NotifyMessage TaskMessage = iota
	InfoMessage
	WarningMessage
	ErrorMessage
)

func (t TaskMessage) Values() []string {
	return []string{
		NotifyMessage.String(),
		InfoMessage.String(),
		WarningMessage.String(),
		ErrorMessage.String(),
	}
}

func (t TaskMessage) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t TaskMessage) String() string {
	switch t {
	case NotifyMessage:
		return "notify"
	case InfoMessage:
		return "info"
	case WarningMessage:
		return "warning"
	case ErrorMessage:
		return "error"
	default:
		return "unknown"
	}
}

func (t *TaskMessage) Scan(val any) error {
	var s string

	switch v := val.(type) {
	case nil:
		return nil
	case []uint8:
		s = string(v)
	case string:
		s = v
	}

	switch s {
	case NotifyMessage.String():
		*t = NotifyMessage
	case InfoMessage.String():
		*t = InfoMessage
	case WarningMessage.String():
		*t = WarningMessage
	case ErrorMessage.String():
		*t = ErrorMessage
	}
	return nil
}
