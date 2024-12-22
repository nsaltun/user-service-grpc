package grpc

import (
	"context"

	grpcserver "github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"google.golang.org/grpc"
)

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//Do auth

		return handler(ctx, req)
	}
}

func WithAuthInterceptor() grpcserver.OptionFn {
	return func(opt *grpcserver.GrpcOption) {
		opt.UnaryInterceptors = append(opt.UnaryInterceptors, AuthInterceptor())
	}
}
