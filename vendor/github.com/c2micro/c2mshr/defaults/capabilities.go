package defaults

import (
	"database/sql/driver"
	"fmt"

	commonv1 "github.com/c2micro/c2mshr/proto/gen/common/v1"
	"google.golang.org/protobuf/proto"
)

type Capability uint32

const (
	CAP_SLEEP               Capability = 0
	CAP_LS                  Capability = 1
	CAP_PWD                 Capability = 2 << 0
	CAP_CD                  Capability = 2 << 1
	CAP_WHOAMI              Capability = 2 << 2
	CAP_PS                  Capability = 2 << 3
	CAP_CAT                 Capability = 2 << 4
	CAP_EXEC                Capability = 2 << 5
	CAP_CP                  Capability = 2 << 6
	CAP_JOBS                Capability = 2 << 7
	CAP_JOBKILL             Capability = 2 << 8
	CAP_KILL                Capability = 2 << 9
	CAP_MV                  Capability = 2 << 10
	CAP_MKDIR               Capability = 2 << 11
	CAP_RM                  Capability = 2 << 12
	CAP_EXEC_ASSEMBLY       Capability = 2 << 13
	CAP_SHELLCODE_INJECTION Capability = 2 << 14
	CAP_DOWNLOAD            Capability = 2 << 15
	CAP_UPLOAD              Capability = 2 << 16
	CAP_PAUSE               Capability = 2 << 17
	CAP_DESTRUCT            Capability = 2 << 18
	CAP_EXEC_DETACH         Capability = 2 << 19
	CAP_SHELL               Capability = 2 << 20
	CAP_PPID                Capability = 2 << 21
	CAP_EXIT                Capability = 2 << 22
	CAP_SOCKS5              Capability = 2 << 23
)

func (c Capability) Values() []string {
	return []string{
		CAP_SLEEP.String(),
		CAP_LS.String(),
		CAP_PWD.String(),
		CAP_CD.String(),
		CAP_WHOAMI.String(),
		CAP_PS.String(),
		CAP_CAT.String(),
		CAP_EXEC.String(),
		CAP_CP.String(),
		CAP_JOBS.String(),
		CAP_JOBKILL.String(),
		CAP_KILL.String(),
		CAP_MV.String(),
		CAP_MKDIR.String(),
		CAP_RM.String(),
		CAP_EXEC_ASSEMBLY.String(),
		CAP_SHELL.String(),
		CAP_PPID.String(),
		CAP_EXEC_DETACH.String(),
		CAP_SHELLCODE_INJECTION.String(),
		CAP_DOWNLOAD.String(),
		CAP_UPLOAD.String(),
		CAP_PAUSE.String(),
		CAP_DESTRUCT.String(),
		CAP_EXIT.String(),
		CAP_SOCKS5.String(),
	}
}

func (c Capability) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c Capability) String() string {
	switch c {
	case CAP_SLEEP:
		return "c_sleep"
	case CAP_LS:
		return "c_ls"
	case CAP_PWD:
		return "c_pwd"
	case CAP_CD:
		return "c_cd"
	case CAP_WHOAMI:
		return "c_whoami"
	case CAP_PS:
		return "c_ps"
	case CAP_CAT:
		return "c_cat"
	case CAP_EXEC:
		return "c_exec"
	case CAP_CP:
		return "c_cp"
	case CAP_JOBS:
		return "c_jobs"
	case CAP_JOBKILL:
		return "c_jobkill"
	case CAP_KILL:
		return "c_kill"
	case CAP_MV:
		return "c_mv"
	case CAP_MKDIR:
		return "c_mkdir"
	case CAP_RM:
		return "c_rm"
	case CAP_EXEC_ASSEMBLY:
		return "c_exec_assembly"
	case CAP_SHELL:
		return "c_shell"
	case CAP_PPID:
		return "c_ppid"
	case CAP_EXEC_DETACH:
		return "c_exec_detach"
	case CAP_SHELLCODE_INJECTION:
		return "c_shellcode_injection"
	case CAP_DOWNLOAD:
		return "c_download"
	case CAP_UPLOAD:
		return "c_upload"
	case CAP_PAUSE:
		return "c_pause"
	case CAP_DESTRUCT:
		return "c_destruct"
	case CAP_EXIT:
		return "c_exit"
	case CAP_SOCKS5:
		return "c_socks5"
	default:
		return "unknown"
	}
}

func (c *Capability) Scan(val any) error {
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
	case CAP_SLEEP.String():
		*c = CAP_SLEEP
	case CAP_LS.String():
		*c = CAP_LS
	case CAP_PWD.String():
		*c = CAP_PWD
	case CAP_CD.String():
		*c = CAP_CD
	case CAP_WHOAMI.String():
		*c = CAP_WHOAMI
	case CAP_PS.String():
		*c = CAP_PS
	case CAP_CAT.String():
		*c = CAP_CAT
	case CAP_EXEC.String():
		*c = CAP_EXEC
	case CAP_CP.String():
		*c = CAP_CP
	case CAP_JOBS.String():
		*c = CAP_JOBS
	case CAP_JOBKILL.String():
		*c = CAP_JOBKILL
	case CAP_KILL.String():
		*c = CAP_KILL
	case CAP_MV.String():
		*c = CAP_MV
	case CAP_MKDIR.String():
		*c = CAP_MKDIR
	case CAP_RM.String():
		*c = CAP_RM
	case CAP_EXEC_ASSEMBLY.String():
		*c = CAP_EXEC_ASSEMBLY
	case CAP_SHELL.String():
		*c = CAP_SHELL
	case CAP_PPID.String():
		*c = CAP_PPID
	case CAP_EXEC_DETACH.String():
		*c = CAP_EXEC_DETACH
	case CAP_SHELLCODE_INJECTION.String():
		*c = CAP_SHELLCODE_INJECTION
	case CAP_DOWNLOAD.String():
		*c = CAP_DOWNLOAD
	case CAP_UPLOAD.String():
		*c = CAP_UPLOAD
	case CAP_PAUSE.String():
		*c = CAP_PAUSE
	case CAP_DESTRUCT.String():
		*c = CAP_DESTRUCT
	case CAP_EXIT.String():
		*c = CAP_EXIT
	case CAP_SOCKS5.String():
		*c = CAP_SOCKS5
	}
	return nil
}

// Marshal маршалим прото сообщение, в зависимости от типа капы
func (c Capability) Marshal(data any) ([]byte, error) {
	switch c {
	case CAP_SLEEP:
		v, ok := data.(*commonv1.CapSleep)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_LS:
		v, ok := data.(*commonv1.CapLs)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_PWD:
		v, ok := data.(*commonv1.CapPwd)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_CD:
		v, ok := data.(*commonv1.CapCd)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_WHOAMI:
		v, ok := data.(*commonv1.CapWhoami)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_PS:
		v, ok := data.(*commonv1.CapPs)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_CAT:
		v, ok := data.(*commonv1.CapCat)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_EXEC:
		v, ok := data.(*commonv1.CapExec)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_CP:
		v, ok := data.(*commonv1.CapCp)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_JOBS:
		v, ok := data.(*commonv1.CapJobs)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_JOBKILL:
		v, ok := data.(*commonv1.CapJobkill)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_KILL:
		v, ok := data.(*commonv1.CapKill)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_MV:
		v, ok := data.(*commonv1.CapMv)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_MKDIR:
		v, ok := data.(*commonv1.CapMkdir)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_RM:
		v, ok := data.(*commonv1.CapRm)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_EXEC_ASSEMBLY:
		v, ok := data.(*commonv1.CapExecAssembly)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_SHELL:
		v, ok := data.(*commonv1.CapShell)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_PPID:
		v, ok := data.(*commonv1.CapPpid)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_EXEC_DETACH:
		v, ok := data.(*commonv1.CapExecDetach)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_SHELLCODE_INJECTION:
		v, ok := data.(*commonv1.CapShellcodeInjection)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_DOWNLOAD:
		v, ok := data.(*commonv1.CapDownload)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_UPLOAD:
		v, ok := data.(*commonv1.CapUpload)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_PAUSE:
		v, ok := data.(*commonv1.CapPause)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_DESTRUCT:
		v, ok := data.(*commonv1.CapDestruct)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_EXIT:
		v, ok := data.(*commonv1.CapExit)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	case CAP_SOCKS5:
		v, ok := data.(*commonv1.CapSocks5)
		if !ok {
			return nil, fmt.Errorf("%s: invalid argument to marshal", c.String())
		}
		return proto.Marshal(v)
	default:
		return nil, fmt.Errorf("%s: unknown capability type to marshal", c.String())
	}
}

// анмаршалинг массива байт в прото сообщение
func (c Capability) Unmarshal(data []byte) (any, error) {
	switch c {
	case CAP_SLEEP:
		v := new(commonv1.CapSleep)
		return v, proto.Unmarshal(data, v)
	case CAP_LS:
		v := new(commonv1.CapLs)
		return v, proto.Unmarshal(data, v)
	case CAP_PWD:
		v := new(commonv1.CapPwd)
		return v, proto.Unmarshal(data, v)
	case CAP_CD:
		v := new(commonv1.CapCd)
		return v, proto.Unmarshal(data, v)
	case CAP_WHOAMI:
		v := new(commonv1.CapWhoami)
		return v, proto.Unmarshal(data, v)
	case CAP_PS:
		v := new(commonv1.CapPs)
		return v, proto.Unmarshal(data, v)
	case CAP_CAT:
		v := new(commonv1.CapCat)
		return v, proto.Unmarshal(data, v)
	case CAP_EXEC:
		v := new(commonv1.CapExec)
		return v, proto.Unmarshal(data, v)
	case CAP_CP:
		v := new(commonv1.CapCp)
		return v, proto.Unmarshal(data, v)
	case CAP_JOBS:
		v := new(commonv1.CapJobs)
		return v, proto.Unmarshal(data, v)
	case CAP_JOBKILL:
		v := new(commonv1.CapJobkill)
		return v, proto.Unmarshal(data, v)
	case CAP_KILL:
		v := new(commonv1.CapKill)
		return v, proto.Unmarshal(data, v)
	case CAP_MV:
		v := new(commonv1.CapMv)
		return v, proto.Unmarshal(data, v)
	case CAP_MKDIR:
		v := new(commonv1.CapMkdir)
		return v, proto.Unmarshal(data, v)
	case CAP_RM:
		v := new(commonv1.CapRm)
		return v, proto.Unmarshal(data, v)
	case CAP_EXEC_ASSEMBLY:
		v := new(commonv1.CapExecAssembly)
		return v, proto.Unmarshal(data, v)
	case CAP_SHELL:
		v := new(commonv1.CapShell)
		return v, proto.Unmarshal(data, v)
	case CAP_PPID:
		v := new(commonv1.CapPpid)
		return v, proto.Unmarshal(data, v)
	case CAP_EXEC_DETACH:
		v := new(commonv1.CapExecDetach)
		return v, proto.Unmarshal(data, v)
	case CAP_SHELLCODE_INJECTION:
		v := new(commonv1.CapShellcodeInjection)
		return v, proto.Unmarshal(data, v)
	case CAP_DOWNLOAD:
		v := new(commonv1.CapDownload)
		return v, proto.Unmarshal(data, v)
	case CAP_UPLOAD:
		v := new(commonv1.CapUpload)
		return v, proto.Unmarshal(data, v)
	case CAP_PAUSE:
		v := new(commonv1.CapPause)
		return v, proto.Unmarshal(data, v)
	case CAP_DESTRUCT:
		v := new(commonv1.CapDestruct)
		return v, proto.Unmarshal(data, v)
	case CAP_EXIT:
		v := new(commonv1.CapExit)
		return v, proto.Unmarshal(data, v)
	case CAP_SOCKS5:
		v := new(commonv1.CapSocks5)
		return v, proto.Unmarshal(data, v)
	default:
		return nil, fmt.Errorf("%s: unknown capability type to unmarshal", c.String())
	}
}

// функция для валидации маски с капами
// возвращает true, если маска содержит капу
func (c Capability) ValidateMask(cap uint32) bool {
	return cap&uint32(c) == uint32(c)
}

// получение списка капов на базе маски
func SupportedCaps(mask uint32) []Capability {
	var t []Capability
	if mask&uint32(CAP_SLEEP) == uint32(CAP_SLEEP) {
		t = append(t, CAP_SLEEP)
	}
	if mask&uint32(CAP_LS) == uint32(CAP_LS) {
		t = append(t, CAP_LS)
	}
	if mask&uint32(CAP_PWD) == uint32(CAP_PWD) {
		t = append(t, CAP_PWD)
	}
	if mask&uint32(CAP_CD) == uint32(CAP_CD) {
		t = append(t, CAP_CD)
	}
	if mask&uint32(CAP_WHOAMI) == uint32(CAP_WHOAMI) {
		t = append(t, CAP_WHOAMI)
	}
	if mask&uint32(CAP_PS) == uint32(CAP_PS) {
		t = append(t, CAP_PS)
	}
	if mask&uint32(CAP_CAT) == uint32(CAP_CAT) {
		t = append(t, CAP_CAT)
	}
	if mask&uint32(CAP_EXEC) == uint32(CAP_EXEC) {
		t = append(t, CAP_EXEC)
	}
	if mask&uint32(CAP_CP) == uint32(CAP_CP) {
		t = append(t, CAP_CP)
	}
	if mask&uint32(CAP_JOBS) == uint32(CAP_JOBS) {
		t = append(t, CAP_JOBS)
	}
	if mask&uint32(CAP_JOBKILL) == uint32(CAP_JOBKILL) {
		t = append(t, CAP_JOBKILL)
	}
	if mask&uint32(CAP_KILL) == uint32(CAP_KILL) {
		t = append(t, CAP_KILL)
	}
	if mask&uint32(CAP_MV) == uint32(CAP_MV) {
		t = append(t, CAP_MV)
	}
	if mask&uint32(CAP_MKDIR) == uint32(CAP_MKDIR) {
		t = append(t, CAP_MKDIR)
	}
	if mask&uint32(CAP_RM) == uint32(CAP_RM) {
		t = append(t, CAP_RM)
	}
	if mask&uint32(CAP_EXEC_ASSEMBLY) == uint32(CAP_EXEC_ASSEMBLY) {
		t = append(t, CAP_EXEC_ASSEMBLY)
	}
	if mask&uint32(CAP_SHELLCODE_INJECTION) == uint32(CAP_SHELLCODE_INJECTION) {
		t = append(t, CAP_SHELLCODE_INJECTION)
	}
	if mask&uint32(CAP_DOWNLOAD) == uint32(CAP_DOWNLOAD) {
		t = append(t, CAP_DOWNLOAD)
	}
	if mask&uint32(CAP_UPLOAD) == uint32(CAP_UPLOAD) {
		t = append(t, CAP_UPLOAD)
	}
	if mask&uint32(CAP_PAUSE) == uint32(CAP_PAUSE) {
		t = append(t, CAP_PAUSE)
	}
	if mask&uint32(CAP_DESTRUCT) == uint32(CAP_DESTRUCT) {
		t = append(t, CAP_DESTRUCT)
	}
	if mask&uint32(CAP_EXEC_DETACH) == uint32(CAP_EXEC_DETACH) {
		t = append(t, CAP_EXEC_DETACH)
	}
	if mask&uint32(CAP_SHELL) == uint32(CAP_SHELL) {
		t = append(t, CAP_SHELL)
	}
	if mask&uint32(CAP_PPID) == uint32(CAP_PPID) {
		t = append(t, CAP_PPID)
	}
	if mask&uint32(CAP_EXIT) == uint32(CAP_EXIT) {
		t = append(t, CAP_EXIT)
	}
	if mask&uint32(CAP_SOCKS5) == uint32(CAP_SOCKS5) {
		t = append(t, CAP_SOCKS5)
	}
	return t
}
