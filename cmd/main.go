package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsaltun/user-service-grpc/internal/service"
	"github.com/nsaltun/user-service-grpc/pkg/v1/logging"
	"github.com/nsaltun/user-service-grpc/proto/generated/go/userapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logging.InitSlog()
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		slog.Error("Failed to listen TCP port", "port", "3000", "err", err)
		os.Exit(1)
	}

	server := grpc.NewServer()

	//Register userapi to server
	userService := service.NewUserAPI()
	userapi.RegisterUserAPIServer(server, userService)
	reflection.Register(server)

	go func() {
		if err := server.Serve(listener); err != nil {
			slog.Error("grpc server serve error.", "err", err)
			os.Exit(1)
		}
	}()
	slog.Info("Grpc server is running", "address", "tcp:3000")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Stopping grpc server..")
	server.GracefulStop()
}
