package main

import (
	"github.com/nsaltun/user-service-grpc/internal/repository"
	"github.com/nsaltun/user-service-grpc/internal/service"
	"github.com/nsaltun/user-service-grpc/pkg/v1/db/mongohandler"
	"github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"github.com/nsaltun/user-service-grpc/pkg/v1/logging"
	grpcmiddl "github.com/nsaltun/user-service-grpc/pkg/v1/middleware/grpc"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	userapi "github.com/nsaltun/user-service-grpc/proto/gen/go/core/user/v1"
)

func main() {
	s := stack.New()
	defer s.Close()

	logging.InitSlog()

	// Init mongodb
	mongoWrapper := mongohandler.New()
	s.MustInit(mongoWrapper)

	// Init repository
	userRepo := repository.NewUserRepo(mongoWrapper)
	s.MustInit(userRepo)

	// Register userapi to server
	userService := service.NewUserAPI(repository.New(userRepo))
	s.MustInit(userService)

	// grpc server
	grpcServer := grpc.New(
		grpcmiddl.WithAuthInterceptor(),
		grpcmiddl.WithLoggingInterceptor(),
		grpcmiddl.WithErrorInterceptor(), //error interceptor must be the last one
	)
	userapi.RegisterUserAPIServer(grpcServer.Server(), userService)

	//grpcServer must init in the end
	s.MustInit(grpcServer)
}
