package grpc

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nsaltun/user-service-grpc/pkg/v1/stack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	runOnce sync.Once
)

type GrpcServer interface {
	stack.Provider
	Server() *grpc.Server
}

type server struct {
	stack.AbstractProvider
	config     ServerConfig
	options    *GrpcOption
	grpcServer *grpc.Server
}

type GrpcOption struct {
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
}
type OptionFn func(*GrpcOption)

func New(options ...OptionFn) GrpcServer {
	grpcOption := &GrpcOption{}
	for _, o := range options {
		o(grpcOption)
	}

	return &server{
		config:     NewServerConfigFromEnv(),
		options:    grpcOption,
		grpcServer: grpc.NewServer(grpc.ChainUnaryInterceptor(grpcOption.UnaryInterceptors...), grpc.ChainStreamInterceptor(grpcOption.StreamInterceptors...)),
	}
}

func (s *server) Init() error {
	reflection.Register(s.grpcServer)

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		slog.Error("Failed to listen TCP port", "port", "3000", "err", err)
		return err
	}

	// RunOnce makes sure the Stack isn't started twice.
	runOnce.Do(func() {
		go func() {
			if err = s.grpcServer.Serve(listener); err != nil {
				slog.Error("grpc server serve error.", "err", err)
				return
			}
		}()
		slog.Info("Grpc server is running", "address", "tcp:3000")
		s.handleInterrupt()
	})

	return nil
}

func (s *server) Server() *grpc.Server {
	return s.grpcServer
}

func (s *server) handleInterrupt() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Stopping grpc server..")
	s.grpcServer.GracefulStop()
}
