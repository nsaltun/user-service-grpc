package service

import (
	"context"

	"github.com/nsaltun/user-service-grpc/proto/generated/go/userapi/v1"
)

type UserAPI interface {
	userapi.UserAPIServer
}

type userService struct {
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
