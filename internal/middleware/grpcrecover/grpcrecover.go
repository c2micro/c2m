package grpcrecover

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// recoverFrom при панике отдает internal ошибку клиенту
func recoverFrom(_ context.Context, p interface{}) error {
	fmt.Printf("%s\n%s", p, debug.Stack())
	return status.Errorf(codes.Internal, "internal server error")
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				err = recoverFrom(ctx, r)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				err = recoverFrom(stream.Context(), r)
			}
		}()

		err = handler(srv, stream)
		panicked = false
		return err
	}
}
