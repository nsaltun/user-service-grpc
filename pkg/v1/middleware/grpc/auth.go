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

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the key used to store the user ID in the context
	UserIDKey contextKey = "user_id"
	// DeviceIDKey is the key used to store the device ID in the context
	DeviceIDKey contextKey = "device_id"
	// TokenFamilyKey is the key used to store the token family ID in the context
	TokenFamilyKey contextKey = "token_family"
)

func AuthInterceptor(jwtManager *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract device information from metadata
		deviceID := extractDeviceID(ctx)

		// Add device ID to context for downstream use
		ctx = context.WithValue(ctx, DeviceIDKey, deviceID)

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
			// Add claims information to context
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
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

// extractDeviceID extracts device information from the request metadata
func extractDeviceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "default"
	}

	// Try to get device ID from metadata
	if values := md.Get("x-device-id"); len(values) > 0 {
		return values[0]
	}

	// Try to get from user agent as fallback
	if values := md.Get("user-agent"); len(values) > 0 {
		return values[0]
	}

	return "default"
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok && userID != ""
}

// GetDeviceID retrieves the device ID from the context
func GetDeviceID(ctx context.Context) (string, bool) {
	deviceID, ok := ctx.Value(DeviceIDKey).(string)
	return deviceID, ok && deviceID != ""
}

// GetTokenFamily retrieves the token family ID from the context
func GetTokenFamily(ctx context.Context) (string, bool) {
	familyID, ok := ctx.Value(TokenFamilyKey).(string)
	return familyID, ok && familyID != ""
}
