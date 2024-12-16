package run

import (
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/listener"
	"github.com/c2micro/c2msrv/internal/management"
	"github.com/c2micro/c2msrv/internal/operator"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type cmdRun struct {
	lg            *zap.Logger
	db            *ent.Client
	operatorCfg   operator.ConfigV1
	listenerCfg   listener.ConfigV1
	managementCfg management.ConfigV1
}

func (cr *cmdRun) Run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	g, ctx := errgroup.WithContext(ctx)

	// management
	g.Go(func() error {
		return management.Serve(ctx, cr.managementCfg, cr.db)
	})

	// listener
	g.Go(func() error {
		return listener.Serve(ctx, cr.listenerCfg, cr.db)
	})

	// operator
	g.Go(func() error {
		return operator.Serve(ctx, cr.operatorCfg, cr.db)
	})

	return g.Wait()
}
