package user

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/crypt"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	typesv1 "github.com/nsaltun/user-service-grpc/proto/gen/go/shared/types/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type UserService interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	UpdateUserById(ctx context.Context, id string, user *model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsersByFilter(ctx context.Context, filter *model.UserFilter) (*pb.ListUsersResponse, error)
}

type user struct {
	repo repository.Repository
}

func NewUserService(repo repository.Repository) UserService {
	return &user{
		repo: repo,
	}
}

// User service implementations
func (s *user) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	hashedPwd, err := crypt.HashPassword(user.Password)
	if err != nil {
		if err == bcrypt.ErrPasswordTooLong {
			return nil, errwrap.NewError("password is too long", codes.InvalidArgument.String()).SetGrpcCode(codes.InvalidArgument)
		}
		return nil, errwrap.NewError("unexpected error", codes.Internal.String()).SetGrpcCode(codes.Internal).SetOriginError(err)
	}

	user.Password = hashedPwd

	//Set user init default values
	user.Id = uuid.NewString() // Generate a new UUID
	user.Status = model.UserStatus_Active
	user.Meta = types.NewMeta()
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserById updates a user by their ID with partial updates
func (s *user) UpdateUserById(ctx context.Context, id string, user *model.User) (*model.User, error) {
	// Get existing user to check if exists and merge updates
	existingUser, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		return nil, err // Repository should already return appropriate error
	}

	// Update only provided fields (partial update)
	err = applyPartialUpdates(existingUser, *user)
	if err != nil {
		return nil, err
	}

	// Update metadata
	existingUser.Meta.Update()

	// Save updated user
	if err := s.repo.UpdateUser(ctx, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *user) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists
	existingUser, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		return err // Repository should already return appropriate error
	}
	if existingUser.Status == model.UserStatus_Inactive {
		return errwrap.NewError("user is already inactive", codes.InvalidArgument.String()).SetGrpcCode(codes.InvalidArgument)
	}

	// Soft delete by updating status to Deleted
	existingUser.Status = model.UserStatus_Inactive
	existingUser.Meta.Update()

	// Save updated user
	if err := s.repo.UpdateUser(ctx, existingUser); err != nil {
		return err
	}

	return nil
}

func (s *user) ListUsersByFilter(ctx context.Context, filter *model.UserFilter) (*pb.ListUsersResponse, error) {
	filterBsonMap := filter.ToBson()

	// Get users from repository with filter
	users, total, err := s.repo.ListUsers(ctx, filterBsonMap, filter.Pagination)
	if err != nil {
		return nil, err
	}

	// Convert model users to proto users more efficiently
	pbUsers := make([]*pb.User, 0, len(users)) // Pre-allocate with capacity
	for _, user := range users {
		pbUsers = append(pbUsers, user.UserToProto())
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Params: &typesv1.Pagination{
			TotalRecords:  total,
			HasNext:       (filter.Pagination.Offset + filter.Pagination.Limit) < total,
			HasPrevious:   filter.Pagination.Offset > 0,
			CurrentLimit:  filter.Pagination.Limit,
			CurrentOffset: filter.Pagination.Offset,
		},
	}, nil
}

// applyPartialUpdates updates only provided fields from source to target user
func applyPartialUpdates(existingUser *model.User, user model.User) error {
	if user.FirstName != "" {
		existingUser.FirstName = user.FirstName
	}
	if user.LastName != "" {
		existingUser.LastName = user.LastName
	}
	if user.NickName != "" {
		existingUser.NickName = user.NickName
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Country != "" {
		existingUser.Country = user.Country
	}
	if user.Status != model.UserStatus_Unspecified {
		existingUser.Status = user.Status
	}
	if user.Password != "" {
		// Hash new password if provided
		hashedPwd, err := crypt.HashPassword(user.Password)
		if err != nil {
			if err == bcrypt.ErrPasswordTooLong {
				return errwrap.NewError("password is too long", codes.InvalidArgument.String()).SetGrpcCode(codes.InvalidArgument)
			}
			slog.Warn("hash password error", "error", err)
			return errwrap.NewError("unexpected error", codes.Internal.String()).SetGrpcCode(codes.Internal).SetOriginError(err)
		}
		existingUser.Password = hashedPwd
	}
	return nil
}
