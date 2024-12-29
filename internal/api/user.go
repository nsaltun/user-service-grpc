package api

import (
	"context"

	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/internal/service/user"
	"github.com/nsaltun/user-service-grpc/pkg/v1/errwrap"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	"google.golang.org/grpc/codes"
)

type userAPI struct {
	pb.UnimplementedUserAPIServer
	service user.UserService
}

func NewUserAPI(service user.UserService) pb.UserAPIServer {
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

func (a *userAPI) UpdateUserById(ctx context.Context, req *pb.UpdateUserByIdRequest) (*pb.UpdateUserByIdResponse, error) {
	if req.GetId() == "" {
		return nil, errwrap.NewError("user id is required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	if req.GetUser() == nil {
		return nil, errwrap.NewError("user update data is required", codes.InvalidArgument.String()).
			SetGrpcCode(codes.InvalidArgument)
	}

	// Convert proto to model and set ID
	user := &model.User{}
	user.FromProto(req.GetUser())

	// Call service
	updatedUser, err := a.service.UpdateUserById(ctx, req.GetId(), user)
	if err != nil {
		return nil, err
	}

	// Convert back to proto and return
	return &pb.UpdateUserByIdResponse{
		User: updatedUser.ToProto(),
	}, nil
}
