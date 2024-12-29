package service

import (
	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/internal/service/auth"
	"github.com/nsaltun/user-service-grpc/internal/service/user"
)

type Service interface {
	user.UserService
	auth.AuthService
}

type service struct {
	repo repository.Repository
	user.UserService
	auth.AuthService
}

func NewService(repo repository.Repository) Service {
	svc := &service{
		repo: repo,
	}
	svc.UserService = user.NewUserService(repo)
	svc.AuthService = auth.NewAuthService(repo)
	return svc
}
