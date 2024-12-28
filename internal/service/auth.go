package service

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type AuthAPI interface {
	stack.Provider
	pb.AuthServiceServer
}

type authService struct {
	stack.AbstractProvider
	pb.UnimplementedAuthServiceServer
	repo repository.Repository
}

func NewAuthAPI(repo repository.Repository) AuthAPI {
	return &authService{repo: repo}
}

func (s *authService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Input validation
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errwrap.NewError("email and password are required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, errwrap.NewError("user not found", codes.NotFound.String()).
			SetGrpcCode(codes.NotFound)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword())); err != nil {
		return nil, errwrap.NewError("invalid credentials", codes.Unauthenticated.String()).
			SetGrpcCode(codes.Unauthenticated)
	}

	// Return empty response with OK status
	return &pb.LoginResponse{}, nil
}
