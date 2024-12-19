package grpcauth

import (
	"context"
	"strings"
	"time"

	"github.com/c2micro/c2mshr/defaults"
	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/c2micro/c2m/internal/ent"
	"github.com/c2micro/c2m/internal/ent/listener"
	"github.com/c2micro/c2m/internal/ent/operator"
	"github.com/c2micro/c2m/internal/middleware"
	"github.com/c2micro/c2m/internal/pools"
	"github.com/go-faster/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type listenerCtxId struct{}

func ListenerToCtx(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, listenerCtxId{}, id)
}

func ListenerFromCtx(ctx context.Context) int {
	return ctx.Value(listenerCtxId{}).(int)
}

func UnaryServerInterceptorListener(db *ent.ListenerClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthListenerHeader)
		if len(tokens) != 1 {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		v, err := db.Query().
			Where(listener.Token(tokens[0])).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
			} else {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		v, err = v.Update().
			SetLast(time.Now()).
			Save(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// нотификация всех подписчиков об изменении времени последней актвиности оператора
		go pools.Pool.Listeners.Send(&operatorv1.ListenerLastResponse{
			Lid:  int64(v.ID),
			Last: timestamppb.New(v.Last),
		})
		return handler(ListenerToCtx(ctx, v.ID), req)
	}
}

func StreamServerInterceptorListener(db *ent.ListenerClient) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthListenerHeader)
		if len(tokens) != 1 {
			return status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		v, err := db.Query().
			Where(listener.Token(tokens[0])).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return status.Error(codes.Unauthenticated, "unauthenticated request")
			} else {
				return status.Error(codes.Internal, err.Error())
			}
		}
		v, err = v.Update().
			SetLast(time.Now()).
			Save(ctx)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		// нотификация всех подписчиков об изменении времени последней актвиности оператора
		go pools.Pool.Listeners.Send(&operatorv1.ListenerLastResponse{
			Lid:  int64(v.ID),
			Last: timestamppb.New(v.Last),
		})
		return handler(srv, &middleware.SrvStream{ServerStream: ss, Ctx: ListenerToCtx(ctx, v.ID)})
	}
}

type operatorCtxId struct{}

func OperatorToCtx(ctx context.Context, u string) context.Context {
	return context.WithValue(ctx, operatorCtxId{}, u)
}

func OperatorFromCtx(ctx context.Context) string {
	return ctx.Value(operatorCtxId{}).(string)
}

func UnaryServerInterceptorOperator(db *ent.OperatorClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthOperatorHeader)
		if len(tokens) != 1 {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		v, err := db.Query().
			Where(operator.Token(tokens[0])).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
			} else {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		v, err = v.Update().
			SetLast(time.Now()).
			Save(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// нотификация всех подписчиков об изменении времени последней актвиности оператора
		go pools.Pool.Operators.Send(&operatorv1.OperatorLastResponse{
			Username: v.Username,
			Last:     timestamppb.New(v.Last),
		})
		return handler(OperatorToCtx(ctx, v.Username), req)
	}
}

func StreamServerInterceptorOperator(db *ent.OperatorClient) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthOperatorHeader)
		if len(tokens) != 1 {
			return status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		v, err := db.Query().
			Where(operator.Token(tokens[0])).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return status.Error(codes.Unauthenticated, "unauthenticated request")
			} else {
				return status.Error(codes.Internal, errors.Wrap(err, "query operator token").Error())
			}
		}
		v, err = v.Update().
			SetLast(time.Now()).
			Save(ctx)
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "update operator last").Error())
		}
		// нотификация всех подписчиков об изменении времени последней актвиности оператора
		go pools.Pool.Operators.Send(&operatorv1.OperatorLastResponse{
			Username: v.Username,
			Last:     timestamppb.New(v.Last),
		})
		return handler(srv, &middleware.SrvStream{ServerStream: ss, Ctx: OperatorToCtx(ctx, v.Username)})
	}
}

// UnaryServerInterceptorManagement мидлваря для аутентификации запросов в management GRPC сервер
func UnaryServerInterceptorManagement(t string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthManagementHeader)
		if len(tokens) != 1 {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		if strings.Compare(tokens[0], t) != 0 {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptorManagement мидлваря для аутентификации запросов в management GRPC сервер
func StreamServerInterceptorManagement(t string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Internal, "missing metadata")
		}
		tokens := meta.Get(defaults.GrpcAuthManagementHeader)
		if len(tokens) != 1 {
			return status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		if strings.Compare(tokens[0], t) != 0 {
			return status.Error(codes.Unauthenticated, "unauthenticated request")
		}
		return handler(srv, ss)
	}
}
