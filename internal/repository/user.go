package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/pkg/v1/db/mongohandler"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
)

type UserRepo interface {
	stack.Provider
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
	stack.AbstractProvider
	collection *mongo.Collection
}

func NewUserRepo(mongoWrapper *mongohandler.MongoDBWrapper) UserRepo {
	return &userRepository{collection: mongoWrapper.Database.Collection("users")}
}

// Init mongo collection (indexes etc.)
func (r *userRepository) Init() error {
	return r.createIndexes()
}

// createIndexes creates indexes specific to the User collection
//
// Creating index for `email`(unique) and `nickName`(unique) and `country`.
func (r *userRepository) createIndexes() error {
	// Define index models
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}}, // Ascending index on email
			Options: options.Index().SetUnique(true),  // Unique constraint
		},
		{
			Keys:    bson.D{{Key: "country", Value: 1}}, // Ascending index on country
			Options: options.Index(),                    // Background creation
		},
		{
			Keys:    bson.D{{Key: "nick_name", Value: 1}}, // Ascending index on nickName
			Options: options.Index().SetUnique(true),      // Unique constraint
		},
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		slog.ErrorContext(ctx, "Error creating indexes for users collection", slog.Any("error", err))
		return err
	}

	slog.InfoContext(ctx, "Indexes created successfully for users collection.")
	return nil
}

// Create a new user
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.collection.InsertOne(ctx, user)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			slog.InfoContext(ctx, "already exists with the same nickname or email.", slog.Any("error", err))
			return errwrap.ErrConflict.SetMessage("already exists with the same nickname or email")
		}
		slog.ErrorContext(ctx, "mongo create user error", slog.Any("error", err), slog.Any("user", user))
		return errwrap.ErrInternal.SetMessage("internal error").SetOriginError(err)
	}

	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errwrap.NewError("user not found", codes.NotFound.String()).
				SetGrpcCode(codes.NotFound)
		}
		return nil, errwrap.NewError("database error", codes.Internal.String()).
			SetGrpcCode(codes.Internal)
	}

	return &user, nil
}
