package auth

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) error
}

type auth struct {
	repo repository.Repository
}

func NewAuthService(repo repository.Repository) AuthService {
	return &auth{
		repo: repo,
	}
}

func (s *auth) Login(ctx context.Context, email, password string) error {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return errwrap.NewError("user not found", codes.NotFound.String()).SetGrpcCode(codes.NotFound)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return errwrap.NewError("invalid credentials", codes.Unauthenticated.String()).SetGrpcCode(codes.Unauthenticated).SetOriginError(err)
	}

	return nil
}
