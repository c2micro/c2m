package run

import (
	"github.com/c2micro/c2m/internal/cfg"

	"github.com/go-faster/sdk/zctx"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Command() *cobra.Command {
	c := cmdRun{}
	return &cobra.Command{
		Use:               "run",
		Short:             "Run c2m server",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.Run,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			var err error

			ctx := cmd.Context()
			c.lg = zctx.From(ctx).Named("run")
			ctxCfg := cfg.GetConfigCtx(ctx)

			// инициализация БД
			if c.db, err = ctxCfg.Db.Init(ctx); err != nil {
				return err
			}
			// инициализация PKI для GRPC серверов
			if err = ctxCfg.Pki.Init(ctx, c.db, ctxCfg.Listener.IP, ctxCfg.Operator.IP, ctxCfg.Management.IP); err != nil {
				return err
			}

			c.operatorCfg = ctxCfg.Operator
			c.listenerCfg = ctxCfg.Listener
			c.managementCfg = ctxCfg.Management

			// инициализация вебхуков
			ctxCfg.Webhook.Init(ctx)

			return nil
		},
		PostRun: func(cmd *cobra.Command, _ []string) {
			// закрытие кона к БД
			if err := c.db.Close(); err != nil {
				c.lg.Error("closing storage", zap.Error(err))
			}
			c.lg.Debug("storage closed")
		},
	}
}
