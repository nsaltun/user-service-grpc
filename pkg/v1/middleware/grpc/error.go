package grpc

import (
	"context"

	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	grpcserver "github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor handles error mapping from application errors to gRPC errors
func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Call the handler
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Check if the error implements IError interface
		if ierr, ok := err.(errwrap.IError); ok {
			return resp, status.Error(ierr.GrpcCode(), ierr.Error())
		}

		// If error doesn't implement IError, return internal server error
		return resp, status.Error(codes.Internal, "internal server error")
	}
}

// WithErrorInterceptor adds the error interceptor to the gRPC server options
func WithErrorInterceptor() grpcserver.OptionFn {
	return func(opt *grpcserver.GrpcOption) {
		opt.UnaryInterceptors = append(opt.UnaryInterceptors, ErrorInterceptor())
	}
}
