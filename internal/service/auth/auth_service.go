package auth

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/auth"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type auth_service struct {
	repo       repository.Repository
	jwtManager *auth.JWTManager
}

func NewAuthService(repo repository.Repository, jwtManager *auth.JWTManager) AuthService {
	return &auth_service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *auth_service) Login(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errwrap.NewError("user not found", codes.NotFound.String()).SetGrpcCode(codes.NotFound)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errwrap.NewError("invalid credentials", codes.Unauthenticated.String()).SetGrpcCode(codes.Unauthenticated).SetOriginError(err)
	}

	// Generate JWT token
	token, err := s.jwtManager.Generate(ctx, user.Id)

	if err != nil {
		return "", errwrap.ErrInternal.SetMessage("failed to generate token").SetOriginError(err)
	}

	return token, nil
}
