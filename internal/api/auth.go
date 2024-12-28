package api

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/service"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	"google.golang.org/grpc/codes"
)

type authAPI struct {
	pb.UnimplementedAuthServiceServer
	service service.AuthService
}

func NewAuthAPI(service service.AuthService) pb.AuthServiceServer {
	return &authAPI{service: service}
}

func (a *authAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Input validation
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errwrap.NewError("email and password are required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	if err := a.service.Login(ctx, req.GetEmail(), req.GetPassword()); err != nil {
		return nil, err
	}

	return &pb.LoginResponse{}, nil
}
