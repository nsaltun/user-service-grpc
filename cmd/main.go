package main

import (
	"log"
	"net"

	"github.com/nsaltun/user-service-grpc/internal/service"
	"github.com/nsaltun/user-service-grpc/proto/generated/go/userapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("Failed to listen. %v", err)
	}

	server := grpc.NewServer()

	//Register userapi to server
	userService := service.NewUserAPI()
	userapi.RegisterUserAPIServer(server, userService)
	reflection.Register(server)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve. %v", err)
	}

}
