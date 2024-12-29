package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//TODO: consider having interface here for better unit testing(mocking)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	authEnabled    bool
	secretKey      string
	tokenDuration  time.Duration
	protectedRoles map[string][]string
}

type tokenParserFn func(ctx context.Context) (string, error)

func NewJWTManager() *JWTManager {
	vi := viper.New()
	vi.AutomaticEnv()

	vi.SetDefault("JWT_SECRET_KEY", "secret")
	vi.SetDefault("JWT_TOKEN_DURATION", "15m")
	vi.SetDefault("JWT_AUTH_ENABLED", "true")

	secretKey := vi.GetString("JWT_SECRET_KEY")
	tokenDuration := vi.GetDuration("JWT_TOKEN_DURATION")
	authEnabled := vi.GetBool("JWT_AUTH_ENABLED")

	//TODO: get from environment variables
	protectedEndpoints := map[string][]string{
		"/core.user.v1.UserAPI/CreateUser":     {"user"}, // protected
		"/core.user.v1.UserAPI/UpdateUserById": {"user"}, // protected
		"/core.user.v1.UserAPI/DeleteUserById": {"user"}, // protected
		"/core.user.v1.UserAPI/ListUsers":      {"user"}, // protected
	}
	return &JWTManager{
		secretKey:      secretKey,
		tokenDuration:  tokenDuration,
		protectedRoles: protectedEndpoints,
		authEnabled:    authEnabled,
	}
}

func (m *JWTManager) Generate(userID string) (string, error) {
	//TODO: invalidate previous token

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) Authorize(ctx context.Context, endpoint string, tokenParser tokenParserFn) (*Claims, error) {
	if !m.needsAuth(endpoint) {
		return nil, nil
	}

	accessToken, err := tokenParser(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := m.Validate(accessToken)
	if err != nil {
		//TODO: check error type. We might have a custom error type. Consider if it's necessary to introduce custom error type.
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return claims, nil
}

func (m *JWTManager) needsAuth(endpoint string) bool {
	if !m.authEnabled {
		return false
	}

	// If roles are empty, method is public
	_, found := m.protectedRoles[endpoint]
	return found
}
