package cmd

import (
	"context"
	"fmt"

	"github.com/c2micro/c2m/internal/cfg"
	"github.com/c2micro/c2m/internal/constants"
	"github.com/c2micro/c2m/internal/utils"

	"github.com/creasty/defaults"
	"github.com/go-faster/errors"
	"github.com/spf13/pflag"
)

type App struct {
	Globals Globals
}

// глобальные флаги
type Globals struct {
	// путь до конфига
	Config string
	// логгирование в дебаге
	Debug bool
}

// регистрация глобальных флагов
func (g *Globals) RegisterFlags(f *pflag.FlagSet) {
	f.StringVarP(&g.Config, "config", "c", utils.EnvOr(constants.ConfigPathEnvKey, ""),
		fmt.Sprintf("path to config file [env:%s]", constants.ConfigPathEnvKey))
	f.BoolVarP(&g.Debug, "debug", "d", false, "enable debug logging")
}

// валидация глобальных флагов
func (g *Globals) Validate(ctx context.Context) (context.Context, error) {
	// вычитываем конфиг
	c, err := cfg.Read(g.Config)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}
	// проставление дефолтов
	if err = defaults.Set(&c); err != nil {
		return nil, errors.Wrap(err, "set defaults")
	}
	// валидируем конфиг
	if err = c.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "validate config")
	}
	// инжектим конфиг в контекст
	return cfg.SetConfigCtx(ctx, c), nil
}
