package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	server := grpc.NewServer()

	//Register userapi to server

	listener, err := net.Listen("tcp", "50051")
	if err != nil {
		log.Fatalf("Failed to listen. %v", err)
	}

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve. %v", err)
	}

}
