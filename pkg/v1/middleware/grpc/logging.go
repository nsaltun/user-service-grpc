package grpc

import (
	"context"
	"log/slog"

	grpcserver "github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"google.golang.org/grpc"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		slog.InfoContext(ctx, "Handling ", "method", info.FullMethod)
		resp, err := handler(ctx, req)
		slog.InfoContext(ctx, "Finished %s with error: %v", info.FullMethod, err)
		return resp, err
	}
}

func WithLoggingInterceptor() grpcserver.OptionFn {
	return func(opt *grpcserver.GrpcOption) {
		opt.UnaryInterceptors = append(opt.UnaryInterceptors, LoggingInterceptor())
	}
}
