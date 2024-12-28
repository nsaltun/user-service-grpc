package api

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/internal/service"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
)

type userAPI struct {
	pb.UnimplementedUserAPIServer
	service service.UserService
}

func NewUserAPI(service service.UserService) pb.UserAPIServer {
	return &userAPI{service: service}
}

func (a *userAPI) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Convert proto to model
	user := &model.User{}
	user.FromProto(req.GetUser())

	// Call service
	createdUser, err := a.service.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Convert back to proto and return
	return &pb.CreateUserResponse{
		User: createdUser.ToProto(),
	}, nil
}
