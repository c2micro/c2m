package main

import (
	"context"
	"os"
	"os/signal"
	"slices"

	"github.com/c2micro/c2msrv/cmd/c2m/internal/cmd"
	"github.com/c2micro/c2msrv/cmd/c2m/internal/cmd/run"
	"github.com/c2micro/c2msrv/cmd/c2m/internal/cmd/version"

	_ "github.com/c2micro/c2msrv/internal/ent/runtime"
	"github.com/c2micro/c2msrv/internal/zapcfg"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	// создание логгера
	lg, err := zapcfg.New().Build()
	if err != nil {
		panic(err)
	}

	// flush функция для закрытия активных объектов
	flush := func() {
		// игнорируем ошибку: /dev/stderr: invalid argument
		_ = lg.Sync()
	}
	defer flush()

	// замена os.Exit
	exit := func(code int) {
		flush()
		os.Exit(code)
	}

	// выход из паники
	defer func() {
		if r := recover(); r != nil {
			lg.Fatal("recovered from panic", zap.Any("panic", r))
			exit(2)
		}
	}()

	// инициализация приложения
	a := cmd.App{}
	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer cancel()

	root := &cobra.Command{
		SilenceUsage:  true, // не показывать usage при ошибке
		SilenceErrors: true,

		Use:   "c2m",
		Short: "c2micro",
		Long:  "c2micro server",
		Args:  cobra.NoArgs,

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// получаем контекст команды
			cmdCtx := cmd.Context()

			// если в имени команды не содержаться определенные подкоманды -> процессинг глобальных флагов
			if !slices.Contains([]string{
				"version",
				"help",
			}, cmd.Name()) {
				// валидация глобальных флагов
				cmdCtx, err = a.Globals.Validate(cmdCtx)
				if err != nil {
					return err
				}
			}
			// обновление уровня логированя
			if a.Globals.Debug {
				zapcfg.AtomLvl.SetLevel(zap.DebugLevel)
			}
			// обновление контекст для команды
			cmd.SetContext(cmdCtx)
			return nil
		},
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			flush()
		},
	}

	// отключаем автокомплит
	root.CompletionOptions.DisableDefaultCmd = true
	// регистрируем глобальные флаги
	a.Globals.RegisterFlags(root.PersistentFlags())
	// регистрация команд
	root.AddCommand(
		version.Command(),
		run.Command(),
	)

	if err := root.ExecuteContext(ctx); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			err = context.Canceled
		}
		color.Red("%v", err)
		exit(2)
	}
}
