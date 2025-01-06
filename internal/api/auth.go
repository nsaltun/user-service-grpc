package api

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/service/auth"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	middleware "github.com/nsaltun/user-service-grpc/pkg/v1/middleware/grpc"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	"google.golang.org/grpc/codes"
)

type authAPI struct {
	pb.UnimplementedAuthServiceServer
	service auth.AuthService
}

func NewAuthAPI(service auth.AuthService) pb.AuthServiceServer {
	return &authAPI{service: service}
}

func (a *authAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Input validation
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errwrap.NewError("email and password are required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	accessToken, refreshToken, err := a.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authAPI) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	// Input validation
	if req.GetRefreshToken() == "" {
		return nil, errwrap.NewError("refresh token is required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	accessToken, refreshToken, err := a.service.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &pb.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authAPI) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// Get user ID from context using the typed key
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, errwrap.ErrUnauthenticated.SetMessage("unauthorized")
	}

	if err := a.service.Logout(ctx, userID); err != nil {
		return nil, err
	}

	return &pb.LogoutResponse{}, nil
}
