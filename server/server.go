package main

import (
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	pb "module/proto"
)

type Server struct {
	pb.UnimplementedChitChatServiceServer

	address string
	mu      sync.Mutex
}

func (s *Server) StartServer() {
	// Create listener
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and register server
	server := grpc.NewServer()
	pb.RegisterChitChatServiceServer(server, s)

	// Log for transparency
	log.Printf("Auction now listening on %s", s.address)

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	log.Println("Server")

	server := &Server{
		address: "localhost:50051",
	}

	server.StartServer()
}
