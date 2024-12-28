package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/pkg/v1/crypt"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type UserService interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]*model.User, error)
}

// User service implementations
func (s *service) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	hashedPwd, err := crypt.HashPassword(user.Password)
	if err != nil {
		code := codes.Internal
		message := "unexpected error"
		if err == bcrypt.ErrPasswordTooLong {
			code = codes.InvalidArgument
			message = "password is too long"
		}
		return nil, errwrap.NewError(message, code.String()).SetGrpcCode(code)
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

func (s *service) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	// TODO: Implement update logic
	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	// TODO: Implement delete logic
	return nil
}

func (s *service) ListUsers(ctx context.Context) ([]*model.User, error) {
	// TODO: Implement list logic
	return nil, nil
}
