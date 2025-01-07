package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nsaltun/user-service-grpc/pkg/v1/db/mongohandler"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Package level constants
const (
	// BufferTimeForExpiration is added to token expiration time to ensure proper cleanup
	BufferTimeForExpiration = 2 * time.Minute

	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Default configuration keys
	configKeyAccessDuration  = "JWT_ACCESS_TOKEN_DURATION"
	configKeyRefreshDuration = "JWT_REFRESH_TOKEN_DURATION"
	configKeyPrivateKey      = "JWT_PRIVATE_KEY"
	configKeyAuthEnabled     = "JWT_AUTH_ENABLED"
	configKeyCollection      = "MONGODB_COLLECTION"

	// Default duration values
	defaultAccessDuration  = "15m" // 15 minutes
	defaultRefreshDuration = "72h" // 3 days
)

// Claims extends jwt.RegisteredClaims with custom fields for our JWT implementation
type Claims struct {
	// UserID uniquely identifies the token owner
	UserID string `json:"user_id"`

	// TokenType specifies whether this is an access or refresh token
	TokenType string `json:"token_type"`

	// DeviceID tracks which device issued the token (used for refresh tokens)
	DeviceID string `json:"device_id,omitempty"`

	// Embed standard JWT claims (exp, iat, etc)
	jwt.RegisteredClaims
}

// UserInvalidatedToken represents a revoked token in the MongoDB collection
type UserInvalidatedToken struct {
	// UserID of the token owner
	UserID string `bson:"user_id"`

	// DeviceID that issued the token (if applicable)
	DeviceID string `bson:"device_id,omitempty"`

	// TokenType distinguishes between access and refresh tokens
	TokenType string `bson:"token_type"`

	// InvalidatedAt tracks when the token was invalidated
	InvalidatedAt time.Time `bson:"invalidated_at"`

	// ExpiresAt is used by MongoDB's TTL index for automatic cleanup
	ExpiresAt time.Time `bson:"expires_at"`
}

// JWTManager handles JWT token operations including generation, validation, and revocation
type JWTManager struct {
	stack.AbstractProvider

	// Configuration
	authEnabled      bool
	privateKeyBase64 string
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey

	// Token settings
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration

	// Access control
	protectedRoles map[string][]string

	// Storage
	collection *mongo.Collection
}

// tokenParserFn defines a function type for extracting tokens from context
type tokenParserFn func(ctx context.Context) (string, error)

// NewJWTManager creates and configures a new JWTManager instance
func NewJWTManager(mongoWrapper *mongohandler.MongoDBWrapper) *JWTManager {
	vi := viper.New()
	vi.AutomaticEnv()

	// Set default configuration values
	vi.SetDefault(configKeyAccessDuration, defaultAccessDuration)
	vi.SetDefault(configKeyRefreshDuration, defaultRefreshDuration)
	vi.SetDefault(configKeyPrivateKey, "")
	vi.SetDefault(configKeyAuthEnabled, "true")
	vi.SetDefault(configKeyCollection, "user_invalidated_tokens")

	// Define protected endpoints and their required roles
	protectedEndpoints := map[string][]string{
		"/core.user.v1.UserAPI/CreateUser":     {"user"},
		"/core.user.v1.UserAPI/UpdateUserById": {"user"},
		"/core.user.v1.UserAPI/DeleteUserById": {"user"},
		"/core.user.v1.UserAPI/ListUsers":      {"user"},
		"/core.user.v1.AuthAPI/Logout":         {"user"},
	}

	collection := mongoWrapper.Collection(vi.GetString(configKeyCollection))

	return &JWTManager{
		privateKeyBase64:     vi.GetString(configKeyPrivateKey),
		accessTokenDuration:  vi.GetDuration(configKeyAccessDuration),
		refreshTokenDuration: vi.GetDuration(configKeyRefreshDuration),
		protectedRoles:       protectedEndpoints,
		authEnabled:          vi.GetBool(configKeyAuthEnabled),
		collection:           collection,
	}
}

// Init initializes the JWT manager by setting up encryption keys and database indexes
func (m *JWTManager) Init() error {
	// Initialize encryption keys
	if err := m.setKeys(); err != nil {
		return fmt.Errorf("failed to initialize encryption keys: %w", err)
	}

	// Create MongoDB indexes
	if err := m.createIndexes(); err != nil {
		return fmt.Errorf("failed to create MongoDB indexes: %w", err)
	}

	return nil
}

// createIndexes sets up the required MongoDB indexes for token management
func (m *JWTManager) createIndexes() error {
	// Create TTL index for automatic token cleanup
	ttlIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
		Options: &options.IndexOptions{
			ExpireAfterSeconds: new(int32), // Expire immediately after expires_at
		},
	}
	if _, err := m.collection.Indexes().CreateOne(context.Background(), ttlIndex); err != nil {
		return fmt.Errorf("failed to create TTL index: %w", err)
	}

	// Create compound index for efficient token queries
	queryIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "invalidated_at", Value: 1},
		},
	}
	if _, err := m.collection.Indexes().CreateOne(context.Background(), queryIndex); err != nil {
		return fmt.Errorf("failed to create query index: %w", err)
	}

	return nil
}

// setKeys initializes the ECDSA key pair for token signing and verification
func (m *JWTManager) setKeys() error {
	// Skip if authentication is disabled
	if !m.authEnabled {
		return nil
	}

	// Validate private key configuration
	if m.privateKeyBase64 == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY environment variable is not set")
	}

	// Decode base64 private key
	privateKeyPEM, err := base64.StdEncoding.DecodeString(m.privateKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("decoded PEM is empty")
	}

	// Clean and normalize the key string
	cleanKey := strings.Trim(string(privateKeyPEM), "\"")
	cleanKey = strings.ReplaceAll(cleanKey, "\\n", "\n")

	// Parse the PEM block
	block, _ := pem.Decode([]byte(cleanKey))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block: invalid PEM format")
	}

	// Verify key type
	if block.Type != "EC PRIVATE KEY" {
		return fmt.Errorf("expected EC PRIVATE KEY but got %s", block.Type)
	}

	// Parse the ECDSA private key
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Store both private and public keys
	m.privateKey = privateKey
	m.publicKey = &privateKey.PublicKey
	return nil
}

// GenerateTokenPair creates a new pair of access and refresh tokens for a user
func (m *JWTManager) GenerateTokenPair(ctx context.Context, userID string, deviceID string) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	// Generate access token first
	accessToken, err = m.generateAccessToken(userID, now)
	if err != nil {
		return "", "", ErrGenerateAccessTokenFailed.SetOriginErr(err)
	}

	// Generate refresh token
	refreshToken, err = m.generateRefreshToken(userID, deviceID, now)
	if err != nil {
		return "", "", ErrGenerateRefreshTokenFailed.SetOriginErr(err)
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken creates a new access token for the given user
func (m *JWTManager) generateAccessToken(userID string, now time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Audience:  jwt.ClaimStrings(maps.Keys(m.protectedRoles)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(m.privateKey)
}

// generateRefreshToken creates a new refresh token for the given user and device
func (m *JWTManager) generateRefreshToken(userID string, deviceID string, now time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(m.privateKey)
}

// Validate verifies the validity of an access token and returns its claims
func (m *JWTManager) Validate(ctx context.Context, tokenStr string) (*Claims, error) {
	var claims Claims

	// Parse and verify the token signature
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, ErrInvalidToken.SetOriginErr(fmt.Errorf("unexpected token signing method"))
			}
			return m.publicKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired.SetOriginErr(err)
		}
		return nil, ErrTokenJwtParse.SetOriginErr(err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify token type
	if claims.TokenType == "" {
		return nil, ErrTokenTypeNotSpecified
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenTypeExpectedAccess
	}

	// Check if token has been invalidated
	filter := bson.M{
		"$or": []bson.M{
			// Check user-wide invalidation
			{
				"user_id":        claims.UserID,
				"token_type":     TokenTypeAccess,
				"invalidated_at": bson.M{"$gte": claims.IssuedAt.Time},
			},
			// Check device-specific invalidation
			{
				"user_id":        claims.UserID,
				"device_id":      claims.DeviceID,
				"token_type":     TokenTypeAccess,
				"invalidated_at": bson.M{"$gte": claims.IssuedAt.Time},
			},
		},
	}

	var invalidToken UserInvalidatedToken
	err = m.collection.FindOne(ctx, filter).Decode(&invalidToken)

	switch {
	case err == nil:
		return nil, ErrTokenInvalidated
	case err == mongo.ErrNoDocuments:
		// No invalidation record found -> token is valid
		return &claims, nil
	default:
		return nil, ErrTokenStatusVerificationFailed.SetOriginErr(err)
	}
}

// RefreshTokens validates a refresh token and generates a new token pair
func (m *JWTManager) RefreshTokens(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, err error) {
	// Validate the refresh token
	claims, err := m.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	now := time.Now()

	// Generate new access token
	accessToken, err = m.generateAccessToken(claims.UserID, now)
	if err != nil {
		return "", "", ErrGenerateAccessTokenFailed.SetOriginErr(err)
	}

	// Generate new refresh token
	newRefreshToken, err = m.generateRefreshToken(claims.UserID, claims.DeviceID, now)
	if err != nil {
		return "", "", ErrGenerateRefreshTokenFailed.SetOriginErr(err)
	}

	// Invalidate the old refresh token
	invalidatedAt := now.Add(-time.Second)
	if err = m.InvalidateToken(ctx, claims.UserID, claims.DeviceID, TokenTypeRefresh, invalidatedAt); err != nil {
		return "", "", ErrInvalidateTokenFailed.SetOriginErr(err)
	}

	return accessToken, newRefreshToken, nil
}

// validateRefreshToken verifies a refresh token's validity and returns its claims
func (m *JWTManager) validateRefreshToken(ctx context.Context, tokenStr string) (*Claims, error) {
	var claims Claims

	// Parse and verify the token signature
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, ErrInvalidToken.SetOriginErr(fmt.Errorf("unexpected token signing method"))
			}
			return m.publicKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired.SetOriginErr(err)
		}
		return nil, ErrTokenJwtParse.SetOriginErr(err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify token type
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenTypeExpectedRefresh
	}

	// Check if token has been invalidated before issued at time.
	filter := bson.M{
		"$or": []bson.M{
			{
				"user_id":    claims.UserID,
				"token_type": TokenTypeRefresh,
				"invalidated_at": bson.M{
					"$gte": claims.IssuedAt.Time,
				},
			},
			{
				"user_id":    claims.UserID,
				"token_type": TokenTypeRefresh,
				"invalidated_at": bson.M{
					"$gte": claims.IssuedAt.Time,
				},
				"device_id": claims.DeviceID,
			},
		},
	}

	var invalidToken UserInvalidatedToken
	err = m.collection.FindOne(ctx, filter).Decode(&invalidToken)
	if err == nil {
		return nil, ErrTokenInvalidated
	} else if err != mongo.ErrNoDocuments {
		return nil, ErrTokenStatusVerificationFailed.SetOriginErr(err)
	}

	return &claims, nil
}

// InvalidateUserTokens revokes all tokens for a specific user
func (m *JWTManager) InvalidateUserTokens(ctx context.Context, userID string) error {
	now := time.Now()

	// Create invalidation records for both token types
	invalidations := []interface{}{
		UserInvalidatedToken{
			UserID:        userID,
			TokenType:     TokenTypeAccess,
			InvalidatedAt: now,
			ExpiresAt:     now.Add(m.accessTokenDuration + BufferTimeForExpiration),
		},
		UserInvalidatedToken{
			UserID:        userID,
			TokenType:     TokenTypeRefresh,
			InvalidatedAt: now,
			ExpiresAt:     now.Add(m.refreshTokenDuration + BufferTimeForExpiration),
		},
	}

	_, err := m.collection.InsertMany(ctx, invalidations)
	if err != nil {
		return ErrInvalidateTokenFailed.SetOriginErr(err)
	}

	return nil
}

// InvalidateTokensBefore revokes all tokens issued before a specific time
func (m *JWTManager) InvalidateTokensBefore(ctx context.Context, userID string, before time.Time) error {
	invalidToken := UserInvalidatedToken{
		UserID:        userID,
		InvalidatedAt: before,
		ExpiresAt:     before.Add(m.accessTokenDuration + BufferTimeForExpiration),
	}

	_, err := m.collection.InsertOne(ctx, invalidToken)
	if err != nil {
		return ErrInvalidateTokenFailed.SetOriginErr(err)
	}

	return nil
}

// InvalidateToken revokes a specific token
func (m *JWTManager) InvalidateToken(ctx context.Context, userID, deviceID, tokenType string, invalidatedAt time.Time) error {
	expiryDuration := m.accessTokenDuration

	// Set expiry duration based on token type
	if tokenType == TokenTypeRefresh {
		expiryDuration = m.refreshTokenDuration
	}

	invalidToken := UserInvalidatedToken{
		UserID:        userID,
		DeviceID:      deviceID,
		TokenType:     tokenType,
		InvalidatedAt: invalidatedAt,
		ExpiresAt:     invalidatedAt.Add(expiryDuration + BufferTimeForExpiration),
	}

	_, err := m.collection.InsertOne(ctx, invalidToken)
	if err != nil {
		return ErrInvalidateTokenFailed.SetOriginErr(err)
	}

	return nil
}

// InvalidateByDeviceID revokes all tokens for a specific device
func (m *JWTManager) InvalidateByDeviceID(ctx context.Context, userID string, deviceID string) error {
	now := time.Now()

	invalidToken := UserInvalidatedToken{
		UserID:        userID,
		DeviceID:      deviceID,
		InvalidatedAt: now,
		ExpiresAt:     now.Add(m.refreshTokenDuration + BufferTimeForExpiration),
	}

	_, err := m.collection.InsertOne(ctx, invalidToken)
	if err != nil {
		return fmt.Errorf("failed to invalidate device tokens: %w", err)
	}

	return nil
}

// Authorize validates a token and checks if it has permission to access an endpoint
func (m *JWTManager) Authorize(ctx context.Context, endpoint string, tokenParser tokenParserFn) (*Claims, error) {
	// Skip authorization if not required
	if !m.needsAuth(endpoint) {
		return nil, nil
	}

	// Extract token from context
	accessToken, err := tokenParser(ctx)
	if err != nil {
		return nil, err
	}

	// Validate the token
	claims, err := m.Validate(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// needsAuth checks if an endpoint requires authentication
func (m *JWTManager) needsAuth(endpoint string) bool {
	if !m.authEnabled {
		return false
	}

	_, found := m.protectedRoles[endpoint]
	return found
}
