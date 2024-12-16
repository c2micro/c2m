package zapcfg

import (
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// AtomLvl уровень логирования в виде глобального объекта
var AtomLvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)

// consoleColorLevelEncoder is single-character color encoder for zapcore.Level.
func consoleColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(color.New(color.FgHiBlack).Sprint("D"))
	case zapcore.InfoLevel:
		enc.AppendString(color.New(color.FgBlue).Sprint("I"))
	case zapcore.WarnLevel:
		enc.AppendString(color.New(color.FgYellow).Sprint("W"))
	case zapcore.ErrorLevel:
		enc.AppendString(color.New(color.FgRed).Sprint("E"))
	case zapcore.FatalLevel:
		enc.AppendString(color.New(color.FgHiRed).Sprint("F"))
	default:
		enc.AppendString("U")
	}
}

// consoleDeltaEncoder colorfully encodes delta from start in seconds and milliseconds.
func consoleDeltaEncoder(now time.Time) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		duration := t.Sub(now)
		seconds := duration / time.Second
		milliseconds := (duration % time.Second) / time.Millisecond
		secColor := color.New(color.Faint)
		msecColor := color.New(color.FgHiBlack)
		enc.AppendString(secColor.Sprintf("%03d", seconds) + msecColor.Sprintf(".%02d", milliseconds/10))
	}
}

func NewDev() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = consoleColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = consoleDeltaEncoder(time.Now())
	cfg.Level = AtomLvl
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.EncodeName = func(s string, encoder zapcore.PrimitiveArrayEncoder) {
		name := s
		//const maxChars = 16
		//if len(name) > maxChars {
		//	name = name[:maxChars]
		//}
		//format := "%-" + strconv.Itoa(maxChars) + "s"
		encoder.AppendString(color.New(color.FgHiBlue).Sprint(name))
	}
	return cfg
}

// NewProd функция инициализации для prod ready логирования (json)
func NewProd() zap.Config {
	cfg := zap.NewProductionConfig()
	// Not so useful, we have go-faster/errors for this that will
	// show in logs.
	cfg.DisableStacktrace = true
	// Human-readable timestamps so ops can read them.
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Disable sampling.
	cfg.Sampling = nil
	// устанавливаем уровень логирования
	cfg.Level = AtomLvl

	return cfg
}

// New создание конфига логгера
func New() zap.Config {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		// если интерактивный терминал (pty) -> подразумевается дев окружения
		return NewDev()
	}
	return NewProd()
}
