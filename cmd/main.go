package main

import (
	"github.com/nsaltun/user-service-grpc/internal/service"
	"github.com/nsaltun/user-service-grpc/pkg/v1/grpc"
	"github.com/nsaltun/user-service-grpc/pkg/v1/logging"
	grpcmiddl "github.com/nsaltun/user-service-grpc/pkg/v1/middleware/grpc"
	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"github.com/nsaltun/user-service-grpc/proto/generated/go/userapi/v1"
)

func main() {
	s := stack.New()
	defer s.Close()

	logging.InitSlog()
	// Register userapi to server
	userService := service.NewUserAPI()
	s.MustInit(userService)

	grpcServer := grpc.New(grpcmiddl.WithAuthInterceptor(), grpcmiddl.WithLoggingInterceptor())
	userapi.RegisterUserAPIServer(grpcServer.Server(), userService)

	//grpcServer must init in the end
	s.MustInit(grpcServer)
}
