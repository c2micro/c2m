package listener

import (
	"context"
	"fmt"
	"net"

	"github.com/c2micro/c2mshr/defaults"
	listenerv1 "github.com/c2micro/c2mshr/proto/gen/listener/v1"
	"github.com/c2micro/c2msrv/internal/constants"
	"github.com/c2micro/c2msrv/internal/ent"
	p "github.com/c2micro/c2msrv/internal/ent/pki"
	"github.com/c2micro/c2msrv/internal/middleware/grpcauth"
	"github.com/c2micro/c2msrv/internal/middleware/grpclog"
	"github.com/c2micro/c2msrv/internal/middleware/grpcrecover"
	"github.com/c2micro/c2msrv/internal/tls"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Serve grpc-сервер для листенеров
func Serve(ctx context.Context, cfg ConfigV1, db *ent.Client) error {
	lg := zctx.From(ctx).Named("listener")

	tlsOpts, f, err := tls.NewTLSTransport(ctx, db, p.TypeListener)
	if err != nil {
		return err
	}
	// создаем сервер
	srv := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsOpts)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  constants.GrpcKeepaliveTime,
			Timeout:               constants.GrpcKeepaliveTimeout,
			MaxConnectionAgeGrace: constants.GrpcMaxConnAgeGrace,
		}),
		grpc.ChainUnaryInterceptor(
			grpcrecover.UnaryServerInterceptor(),
			grpclog.UnaryServerInterceptor(lg),
			grpcauth.UnaryServerInterceptorListener(db.Listener),
		),
		grpc.ChainStreamInterceptor(
			grpcrecover.StreamServerInterceptor(),
			grpclog.StreamServerInterceptor(lg),
			grpcauth.StreamServerInterceptorListener(db.Listener),
		),
		grpc.MaxRecvMsgSize(defaults.MaxProtobufMessageSize),
		grpc.MaxSendMsgSize(defaults.MaxProtobufMessageSize),
	)
	// цепляем сервис
	listenerv1.RegisterListenerServiceServer(srv, &server{
		db: db,
		lg: lg,
	})
	// создаем листенер
	l, err := net.Listen("tcp", fmt.Sprintf(
		"%s:%d", cfg.IP, cfg.Port))
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		lg.Info("start serving",
			zap.String("ip", cfg.IP.String()),
			zap.Int("port", cfg.Port),
			zap.String("fingerprint", f),
		)
		return srv.Serve(l)
	})

	g.Go(func() error {
		// когда родительский контекст закончен - стопаем сервер
		<-ctx.Done()
		srv.Stop()
		lg.Info("stop serving")
		return nil
	})

	return g.Wait()
}
