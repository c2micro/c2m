package defaults

import (
	"database/sql/driver"
)

type BeaconOS uint8

const (
	UnknownOS BeaconOS = iota
	LinuxOS
	WindowsOS
	MacOS
)

func (b BeaconOS) Values() []string {
	return []string{
		UnknownOS.String(),
		LinuxOS.String(),
		WindowsOS.String(),
		MacOS.String(),
	}
}

func (b BeaconOS) Value() (driver.Value, error) {
	return b.String(), nil
}

func (b *BeaconOS) Scan(val any) error {
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
	case UnknownOS.String():
		*b = UnknownOS
	case LinuxOS.String():
		*b = LinuxOS
	case WindowsOS.String():
		*b = WindowsOS
	case MacOS.String():
		*b = MacOS
	}
	return nil
}

func (b BeaconOS) String() string {
	switch b {
	case LinuxOS:
		return "linux"
	case WindowsOS:
		return "windows"
	case MacOS:
		return "macos"
	default:
		return "unknown"
	}
}

func (b BeaconOS) StringShort() string {
	switch b {
	case LinuxOS:
		return "lin"
	case WindowsOS:
		return "win"
	case MacOS:
		return "mac"
	default:
		return "???"
	}
}
