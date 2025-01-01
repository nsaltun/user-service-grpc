package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nsaltun/user-service-grpc/pkg/v1/db/mongohandler"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
JWT System Overview
------------------
This implements a stateless JWT system with the following features:

Storage:
- Only stores invalidation records
- Automatic cleanup using MongoDB TTL index
- Efficient queries via compound indexes

Invalidation Strategies:
1. InvalidateUserTokens
   - Invalidates all tokens for a user
   - Used for logout scenarios

2. InvalidateTokensBefore
   - Invalidates tokens before a specific time
   - Enables selective invalidation

Scalability Benefits:
- Minimal storage overhead (only tracks invalid tokens)
- Automatic cleanup of expired records
- No need to track valid tokens
- Works seamlessly across multiple service instances
*/

const (
	// BufferTimeForExpiration is the time buffer to add to the token expiration time
	// to ensure that the token is not invalidated too early
	BufferTimeForExpiration = 2 * time.Minute
)

// Claims represents the claims in a JWT token
type Claims struct {
	UserID string `json:"user_id"`
	// IAT is automatically included from RegisteredClaims
	jwt.RegisteredClaims
}

// InvalidToken represents an invalidated token in MongoDB
type InvalidToken struct {
	UserID        string    `bson:"user_id"`
	InvalidatedAt time.Time `bson:"invalidated_at"`
	ExpiresAt     time.Time `bson:"expires_at"` // Used for TTL index
}

// JWTManager is a manager for JWT tokens
type JWTManager struct {
	stack.AbstractProvider
	authEnabled      bool
	privateKeyBase64 string
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey
	tokenDuration    time.Duration
	protectedRoles   map[string][]string
	collection       *mongo.Collection
}

// tokenParserFn is a function type for parsing tokens
type tokenParserFn func(ctx context.Context) (string, error)

// NewJWTManager creates a new JWTManager
func NewJWTManager(mongoWrapper *mongohandler.MongoDBWrapper) *JWTManager {
	vi := viper.New()
	vi.AutomaticEnv()

	vi.SetDefault("JWT_TOKEN_DURATION", "15m")
	vi.SetDefault("JWT_PRIVATE_KEY", "")
	vi.SetDefault("JWT_AUTH_ENABLED", "true")
	vi.SetDefault("MONGODB_COLLECTION", "invalid_tokens")

	protectedEndpoints := map[string][]string{
		"/core.user.v1.UserAPI/CreateUser":     {"user"},
		"/core.user.v1.UserAPI/UpdateUserById": {"user"},
		"/core.user.v1.UserAPI/DeleteUserById": {"user"},
		"/core.user.v1.UserAPI/ListUsers":      {"user"},
		"/core.user.v1.AuthService/Logout":     {"user"},
	}

	collection := mongoWrapper.Collection(vi.GetString("MONGODB_COLLECTION"))

	return &JWTManager{
		privateKeyBase64: vi.GetString("JWT_PRIVATE_KEY"),
		tokenDuration:    vi.GetDuration("JWT_TOKEN_DURATION"),
		protectedRoles:   protectedEndpoints,
		authEnabled:      vi.GetBool("JWT_AUTH_ENABLED"),
		collection:       collection,
	}
}

func (m *JWTManager) Init() error {
	err := m.setKeys()
	if err != nil {
		return err
	}

	// Create TTL index on ExpiresAt field
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
		Options: &options.IndexOptions{
			ExpireAfterSeconds: new(int32), // Expire immediately after expires_at
		},
	}
	if _, err := m.collection.Indexes().CreateOne(context.Background(), indexModel); err != nil {
		return fmt.Errorf("failed to create TTL index: %w", err)
	}

	// Create compound index for efficient queries
	compoundIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "issued_at", Value: 1},
		},
	}
	if _, err := m.collection.Indexes().CreateOne(context.Background(), compoundIndex); err != nil {
		return fmt.Errorf("failed to create compound index: %w", err)
	}

	return nil
}

func (m *JWTManager) setKeys() error {
	if !m.authEnabled {
		return nil
	}

	// Read base64 encoded private key from environment variable
	if m.privateKeyBase64 == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY environment variable is not set")
	}

	privateKeyPEM, err := base64.StdEncoding.DecodeString(m.privateKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	// Add debug logging to check PEM content
	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("decoded PEM is empty")
	}

	// Clean up the private key string by removing quotes and replacing escaped newlines
	cleanKey := strings.Trim(string(privateKeyPEM), "\"")
	cleanKey = strings.ReplaceAll(cleanKey, "\\n", "\n")

	// Parse private key directly since it's already in PEM format
	block, _ := pem.Decode([]byte(cleanKey))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block: invalid PEM format")
	}

	// Verify it's an EC PRIVATE KEY
	if block.Type != "EC PRIVATE KEY" {
		return fmt.Errorf("expected EC PRIVATE KEY but got %s", block.Type)
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	m.privateKey = privateKey
	m.publicKey = &privateKey.PublicKey
	return nil
}

func (m *JWTManager) Generate(ctx context.Context, userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.tokenDuration)

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(m.privateKey)
}

func (m *JWTManager) Validate(ctx context.Context, tokenStr string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return m.publicKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is in invalid_tokens collection.
	// We need to check if the token is in the collection because the token might be invalidated by another instance of the service.
	filter := bson.M{
		"user_id": claims.UserID,
		"invalidated_at": bson.M{
			"$gte": claims.IssuedAt.Time,
		},
	}

	var invalidToken InvalidToken
	err = m.collection.FindOne(ctx, filter).Decode(&invalidToken)

	switch {
	case err == nil:
		// Found an invalidation record -> token is invalid
		return nil, fmt.Errorf("token has been invalidated")
	case err == mongo.ErrNoDocuments:
		// No invalidation record found -> token is valid
		return &claims, nil
	default:
		// Unexpected database error
		return nil, fmt.Errorf("failed to verify token status: %w", err)
	}
}

// InvalidateUserTokens invalidates all tokens for a user by creating a record with current timestamp
func (m *JWTManager) InvalidateUserTokens(ctx context.Context, userID string) error {
	now := time.Now()

	invalidToken := InvalidToken{
		UserID:        userID,
		InvalidatedAt: now,
		ExpiresAt:     now.Add(m.tokenDuration + BufferTimeForExpiration),
	}

	_, err := m.collection.InsertOne(ctx, invalidToken)
	if err != nil {
		return fmt.Errorf("failed to invalidate tokens: %w", err)
	}

	return nil
}

// InvalidateTokensBefore invalidates all tokens issued before a specific time
func (m *JWTManager) InvalidateTokensBefore(ctx context.Context, userID string, before time.Time) error {
	invalidToken := InvalidToken{
		UserID:        userID,
		InvalidatedAt: before,
		ExpiresAt:     before.Add(m.tokenDuration + BufferTimeForExpiration),
	}

	_, err := m.collection.InsertOne(ctx, invalidToken)
	if err != nil {
		return fmt.Errorf("failed to invalidate tokens: %w", err)
	}

	return nil
}

func (m *JWTManager) Authorize(ctx context.Context, endpoint string, tokenParser tokenParserFn) (*Claims, error) {
	if !m.needsAuth(endpoint) {
		return nil, nil
	}

	accessToken, err := tokenParser(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := m.Validate(ctx, accessToken)
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
