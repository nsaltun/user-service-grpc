package service

import "github.com/nsaltun/user-service-grpc/internal/repository"

type Service interface {
	UserService
	AuthService
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}
