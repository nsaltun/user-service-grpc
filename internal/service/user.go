package service

import (
	"context"

	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	userapi "github.com/nsaltun/user-service-grpc/proto/gen/go/user/v1"
)

type UserAPI interface {
	stack.Provider
	userapi.UserAPIServer
}

type userService struct {
	stack.AbstractProvider
	userapi.UnimplementedUserAPIServer
}

func NewUserAPI() UserAPI {
	return &userService{}
}

func (s *userService) CreateUser(context.Context, *userapi.CreateUserRequest) (*userapi.CreateUserResponse, error) {
	return &userapi.CreateUserResponse{
		User: &userapi.User{
			FirstName: "enes",
			LastName:  "altun",
			Email:     "enesaltun.dev@gmail.com",
		},
	}, nil
}
