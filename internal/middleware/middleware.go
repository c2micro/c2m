package middleware

import (
	"context"

	"google.golang.org/grpc"
)

type SrvStream struct {
	grpc.ServerStream
	Ctx context.Context
}

func (s *SrvStream) Context() context.Context {
	return s.Ctx
}
