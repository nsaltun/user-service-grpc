package grpc

import (
	"context"
	"strings"

	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	grpcserver "github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nsaltun/user-service-grpc/pkg/v1/auth"
)

func AuthInterceptor(jwtManager *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		claims, err := jwtManager.Authorize(ctx, info.FullMethod, tokenParser)
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				return nil, err
			}

			switch st.Code() {
			case codes.Unauthenticated:
				return nil, errwrap.ErrUnauthenticated.SetOriginError(err)
			case codes.PermissionDenied:
				return nil, errwrap.ErrPermissionDenied.SetOriginError(err)
			default:
				return nil, err
			}
		}

		if claims != nil {
			// Add claims to context
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
		}

		return handler(ctx, req)
	}
}

func WithAuthInterceptor(jwtManager *auth.JWTManager) grpcserver.OptionFn {
	return func(opt *grpcserver.GrpcOption) {
		opt.UnaryInterceptors = append(opt.UnaryInterceptors, AuthInterceptor(jwtManager))
	}
}

func tokenParser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	if !strings.HasPrefix(accessToken, "Bearer ") {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(accessToken, "Bearer "), nil
}
