package auth

import (
	"context"
	"errors"

	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/auth"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
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

func (s *auth_service) Login(ctx context.Context, email, password string) (string, string, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errwrap.NewError("user not found", codes.NotFound.String()).SetGrpcCode(codes.NotFound)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errwrap.NewError("invalid credentials", codes.Unauthenticated.String()).SetGrpcCode(codes.Unauthenticated).SetOriginError(err)
	}

	// Generate token pair
	// TODO: Implement proper device ID management. For now, use a placeholder
	deviceID := "default"
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(ctx, user.Id, deviceID)
	if err != nil {
		return "", "", errwrap.ErrInternal.SetMessage("failed to generate tokens").SetOriginError(err)
	}

	return accessToken, refreshToken, nil
}

func (s *auth_service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	// Validate refresh token and get new token pair
	accessToken, newRefreshToken, err := s.jwtManager.RefreshTokens(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidateTokenFailed) {
			return "", "", errwrap.ErrInternal.SetMessage("error while refreshing token").SetOriginError(err)
		}
		return "", "", errwrap.ErrUnauthenticated.SetMessage(err.Error()).SetOriginError(err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *auth_service) Logout(ctx context.Context, userID string) error {
	// Invalidate all tokens for the user
	if err := s.jwtManager.InvalidateUserTokens(ctx, userID); err != nil {
		return errwrap.ErrInternal.SetMessage("failed to invalidate tokens").SetOriginError(err)
	}

	return nil
}
