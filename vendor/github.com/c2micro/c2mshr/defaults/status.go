package defaults

import (
	"database/sql/driver"
)

type TaskStatus uint8

const (
	StatusNew TaskStatus = iota
	StatusInProgress
	StatusCancelled
	StatusSuccess
	StatusError
)

func (t TaskStatus) Values() []string {
	return []string{
		StatusNew.String(),
		StatusInProgress.String(),
		StatusCancelled.String(),
		StatusSuccess.String(),
		StatusError.String(),
	}
}

func (t TaskStatus) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t TaskStatus) String() string {
	switch t {
	case StatusNew:
		return "new"
	case StatusInProgress:
		return "in-progress"
	case StatusCancelled:
		return "cancelled"
	case StatusSuccess:
		return "success"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

func (t *TaskStatus) Scan(val any) error {
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
	case StatusNew.String():
		*t = StatusNew
	case StatusInProgress.String():
		*t = StatusInProgress
	case StatusCancelled.String():
		*t = StatusCancelled
	case StatusSuccess.String():
		*t = StatusSuccess
	case StatusError.String():
		*t = StatusError
	}
	return nil
}
