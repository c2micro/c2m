package defaults

import (
	"database/sql/driver"
)

type BeaconArch uint8

const (
	UnknownArch BeaconArch = iota
	X86Arch
	X64Arch
	Arm32Arch
	Arm64Arch
)

func (b BeaconArch) Values() []string {
	return []string{
		UnknownArch.String(),
		X86Arch.String(),
		X64Arch.String(),
		Arm32Arch.String(),
		Arm64Arch.String(),
	}
}

func (b BeaconArch) Value() (driver.Value, error) {
	return b.String(), nil
}

func (b BeaconArch) String() string {
	switch b {
	case UnknownArch:
		return "unknown"
	case X86Arch:
		return "x86"
	case X64Arch:
		return "x64"
	case Arm32Arch:
		return "arm32"
	case Arm64Arch:
		return "arm64"
	default:
		return "unknown"
	}
}

func (b *BeaconArch) Scan(val any) error {
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
	case UnknownArch.String():
		*b = UnknownArch
	case X86Arch.String():
		*b = X86Arch
	case X64Arch.String():
		*b = X64Arch
	case Arm32Arch.String():
		*b = Arm32Arch
	case Arm64Arch.String():
		*b = Arm64Arch
	}
	return nil
}
