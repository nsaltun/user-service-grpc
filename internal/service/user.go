package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nsaltun/user-service-grpc/internal/model"
	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"github.com/nsaltun/user-service-grpc/pkg/v1/types"
	pb "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
	typesv1 "github.com/nsaltun/user-service-grpc/proto/gen/go/shared/types/v1"
)

type UserAPI interface {
	stack.Provider
	pb.UserAPIServer
}

type userService struct {
	stack.AbstractProvider
	pb.UnimplementedUserAPIServer
	repo repository.Repository
}

func NewUserAPI(repo repository.Repository) UserAPI {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := &model.User{}
	user.FromProto(request.GetUser())

	//Set user init default values
	user.Id = uuid.NewString() // Generate a new UUID
	user.Status = model.UserStatus_Active
	user.Meta = types.NewMeta()

	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		//TODO err handling/mapping
		return nil, err
	}

	return &pb.CreateUserResponse{User: user.ToProto()}, nil
}

func (s *userService) UpdateUserById(context.Context, *pb.UpdateUserByIdRequest) (*pb.UpdateUserByIdResponse, error) {
	return &pb.UpdateUserByIdResponse{
		User: &pb.User{
			FirstName: "updated firstname",
			LastName:  "updated lastname",
			Email:     "updated@email.com",
		},
	}, nil
}

func (s *userService) DeleteUserById(context.Context, *pb.DeleteUserByIdRequest) (*pb.DeleteUserByIdResponse, error) {
	return &pb.DeleteUserByIdResponse{}, nil
}

func (s *userService) ListUsers(context.Context, *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return &pb.ListUsersResponse{
		Users: []*pb.User{
			{
				Id:        uuid.NewString(),
				FirstName: "user 1 first name",
				LastName:  "user1 last name",
				Email:     "user1@email.com",
			},
			{
				Id:        uuid.NewString(),
				FirstName: "user 2 first name",
				LastName:  "user2 last name",
				Email:     "user2@email.com",
			},
		},
		Params: &typesv1.Pagination{
			TotalRecords:  5,
			CurrentLimit:  2,
			CurrentOffset: 0,
			HasNext:       true,
			HasPrevious:   false,
		},
	}, nil
}
