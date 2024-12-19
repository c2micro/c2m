package grpclog

import (
	"context"
	"time"

	"github.com/c2micro/c2m/internal/middleware"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DefaultCodeToLevel is the defaults implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK:
		return zap.DebugLevel
	case codes.Canceled:
		return zap.InfoLevel
	case codes.Unknown:
		return zap.ErrorLevel
	case codes.InvalidArgument:
		return zap.WarnLevel
	case codes.DeadlineExceeded:
		return zap.WarnLevel
	case codes.NotFound:
		return zap.WarnLevel
	case codes.AlreadyExists:
		return zap.WarnLevel
	case codes.PermissionDenied:
		return zap.WarnLevel
	case codes.Unauthenticated:
		return zap.WarnLevel
	case codes.ResourceExhausted:
		return zap.WarnLevel
	case codes.FailedPrecondition:
		return zap.WarnLevel
	case codes.Aborted:
		return zap.WarnLevel
	case codes.OutOfRange:
		return zap.WarnLevel
	case codes.Unimplemented:
		return zap.ErrorLevel
	case codes.Internal:
		return zap.ErrorLevel
	case codes.Unavailable:
		return zap.WarnLevel
	case codes.DataLoss:
		return zap.ErrorLevel
	default:
		return zap.ErrorLevel
	}
}

func UnaryServerInterceptor(lg *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tStart := time.Now()
		// выполнение запроса
		resp, err := handler(zctx.Base(ctx, lg), req)
		t := time.Since(tStart)
		var code codes.Code
		var msg string
		if s, ok := status.FromError(err); ok {
			code = s.Code()
			msg = s.Message()
		} else {
			code = status.Code(err)
			msg = "<unable extract message from error>"
		}
		go lg.Log(DefaultCodeToLevel(code), "unary call",
			zap.String("method", info.FullMethod),
			zap.String("code", code.String()),
			zap.Duration("time", t),
			zap.String("message", msg),
		)
		return resp, err
	}
}

func StreamServerInterceptor(lg *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		tStart := time.Now()
		// выполнение запроса
		err := handler(srv, &middleware.SrvStream{ServerStream: ss, Ctx: zctx.Base(ss.Context(), lg)})
		t := time.Since(tStart)
		var code codes.Code
		var msg string
		if s, ok := status.FromError(err); ok {
			code = s.Code()
			msg = s.Message()
		} else {
			code = status.Code(err)
			msg = "<unable extract message from error>"
		}
		// лог событие
		go lg.Log(DefaultCodeToLevel(code), "stream call",
			zap.String("method", info.FullMethod),
			zap.String("code", code.String()),
			zap.Duration("time", t),
			zap.String("message", msg),
		)
		return err
	}
}
