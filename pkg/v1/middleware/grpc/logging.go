package grpc

import (
	"context"
	"fmt"
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
		slog.InfoContext(ctx, fmt.Sprintf("Handling method %s", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			slog.InfoContext(ctx, fmt.Sprintf("Finished %s with error", info.FullMethod), "err", err, "resp", resp)
		} else {
			slog.InfoContext(ctx, fmt.Sprintf("Finished %s successfully", info.FullMethod), "resp", resp)
		}
		return resp, err
	}
}

func WithLoggingInterceptor() grpcserver.OptionFn {
	return func(opt *grpcserver.GrpcOption) {
		opt.UnaryInterceptors = append(opt.UnaryInterceptors, LoggingInterceptor())
	}
}
