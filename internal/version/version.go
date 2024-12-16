package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/docker/go-units"

	"github.com/fatih/color"
)

const (
	unknown        = "unknown"
	unknownVersion = "0.0.0"
)

const (
	// commitLength максимальная длина sha1 хэша коммита для вывода
	commitLength = 10
)

// базовая информация по версиям
// используется, если в процессе сборки через ldflags
// не передаются необходимые значения
var (
	gitVersion = unknownVersion
	// gitCommit sha1 хэш коммита
	gitCommit = unknown
	// gitTreeState состояние дерева гита (clean/dirty)
	gitTreeState = unknown
	// buildDate время сборки в ISO8601 формате
	buildDate = unknown
)

// Info структура для хранения информации по версионированию
type Info struct {
	GitVersion   string    `json:"gitVersion"`
	GitCommit    string    `json:"gitCommit"`
	GitTreeState string    `json:"gitTreeState"`
	GitTime      time.Time `json:"gitTime"`
	BuildDate    string    `json:"buildDate"`
	GoVersion    string    `json:"goVersion"`
	Compiler     string    `json:"compiler"`
	Platform     string    `json:"platform"`
	Race         bool      `json:"race"`
}

// Get возвращает инициализированную структуру версионирования
func Get() Info {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// если не удается в рантайме получить билд инфо -> оставляем пустую структуру
		info = &debug.BuildInfo{}
	}

	var vcsTime time.Time
	for _, v := range info.Settings {
		switch v.Key {
		case "vcs.time":
			var err error
			vcsTime, err = time.Parse(time.RFC3339, v.Value)
			if err != nil {
				vcsTime = time.Time{}
			}
		case "vcs.commit":
			gitCommit = v.Value
		}
	}

	// <system>/<arch>
	var p strings.Builder
	p.WriteString(runtime.GOOS)
	p.WriteRune('/')
	p.WriteString(runtime.GOARCH)

	return Info{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GitTime:      vcsTime,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     p.String(),
		Race:         isRace,
	}
}

// Version возвращает строку с версией сервера
func Version() string {
	return gitVersion
}

// String возвращает структуру в виде Go-syntax
func (i Info) String() string {
	return fmt.Sprintf("%#v", i)
}

// Pretty возвращает строку с данными из структуры
func (i Info) Pretty() string {
	var b strings.Builder
	// версия golang
	b.WriteString(i.GoVersion)
	b.WriteRune(' ')
	// платформа
	b.WriteString(i.Platform)
	if i.Race {
		// сервер собран с флагом -race
		b.WriteRune(' ')
		b.WriteString("(race detector enabled)")
	}
	if len(i.GitCommit) > commitLength {
		// хэш коммита
		b.WriteRune(' ')
		b.WriteString(i.GitCommit[:commitLength])
	}
	if i.GitVersion != unknownVersion {
		// версия (берется из тэга)
		b.WriteRune(' ')
		b.WriteString(i.GitVersion)
	}
	return b.String()
}

// PrettyColorful возвращает строку с информацией о версии в цвете
func (i Info) PrettyColorful() string {
	// готовим строку с sha1 хэшом коммита
	c := i.GitCommit
	if len(c) > commitLength {
		c = c[:commitLength]
	}

	var x strings.Builder
	// версия из тега
	x.WriteString(color.New(color.FgGreen, color.Bold).Sprint(i.GitVersion))
	x.WriteRune(' ')
	// версия Go и платформа
	x.WriteString(color.New(color.FgCyan, color.Faint).Sprint(fmt.Sprintf("(%s %s)",
		i.GoVersion, i.Platform)))
	x.WriteRune(' ')
	x.WriteString(color.New(color.FgGreen, color.Faint).Sprint(c))

	if !i.GitTime.IsZero() {
		// если время последнего коммита != 0
		x.WriteRune(' ')
		x.WriteRune('[')
		x.WriteString(color.New(color.FgMagenta, color.Faint).Sprint(
			units.HumanDuration(time.Since(i.GitTime)), " ago"))
		x.WriteRune(']')
	}

	return x.String()
}
