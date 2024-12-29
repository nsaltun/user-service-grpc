package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/pkg/v1/db/mongohandler"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
)

type UserRepo interface {
	stack.Provider
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserById(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	ListUsers(ctx context.Context, filterCriteria bson.M, filter types.PaginationReq) ([]*model.User, int64, error)
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
			SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	return &user, nil
}

func (r *userRepository) GetUserById(ctx context.Context, id string) (*model.User, error) {
	var user model.User

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errwrap.NewError("user not found", codes.NotFound.String()).
				SetGrpcCode(codes.NotFound)
		}
		return nil, errwrap.NewError("database error", codes.Internal.String()).
			SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *model.User) error {
	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": user.Id},
		user,
	)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errwrap.NewError("email or nickname already exists", codes.AlreadyExists.String()).
				SetGrpcCode(codes.AlreadyExists).SetOriginError(err)
		}
		return errwrap.NewError("database error", codes.Internal.String()).
			SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	if result.MatchedCount == 0 {
		return errwrap.NewError("user not found", codes.NotFound.String()).
			SetGrpcCode(codes.NotFound)
	}

	return nil
}

func (r *userRepository) ListUsers(ctx context.Context, filterCriteria bson.M, filter types.PaginationReq) ([]*model.User, int64, error) {
	// Get total count
	total, err := r.collection.CountDocuments(ctx, filterCriteria)
	if err != nil {
		slog.WarnContext(ctx, "mongo list users count error", slog.Any("error", err), slog.Any("filterCriteria", filterCriteria), slog.Any("pagination", filter))
		return nil, 0, errwrap.NewError("database error", codes.Internal.String()).SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	// Create find options for pagination
	findOptions := options.Find()
	findOptions.SetLimit(filter.Limit)
	findOptions.SetSkip(filter.Offset)

	// Execute find with filter
	cursor, err := r.collection.Find(ctx, filterCriteria, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, 0, errwrap.ErrNotFound.SetMessage("user record not found")
		}
		slog.WarnContext(ctx, "mongo list users find error", slog.Any("error", err), slog.Any("filterCriteria", filterCriteria), slog.Any("pagination", filter))
		return nil, 0, errwrap.NewError("database error", codes.Internal.String()).SetGrpcCode(codes.Internal).SetOriginError(err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		slog.WarnContext(ctx, "mongo list users decode error", slog.Any("error", err), slog.Any("filterCriteria", filterCriteria), slog.Any("pagination", filter))
		return nil, 0, errwrap.NewError("database error", codes.Internal.String()).SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	return users, total, nil
}
