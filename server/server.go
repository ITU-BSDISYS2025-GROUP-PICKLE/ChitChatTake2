package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"

	pb "module/proto"
)

type Server struct {
	pb.UnimplementedChitChatServer

	mu            sync.Mutex
	address       string
	clientStreams []pb.ChitChat_JoinServer
	lamportTime   int
}

func (s *Server) StartServer() {
	// Create listener
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and register server
	server := grpc.NewServer()
	pb.RegisterChitChatServer(server, s)

	// Log for transparency
	log.Printf("ChitChat Server now listening on %s", s.address)

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// RPC function. Runs for every client that joins the server
func (s *Server) Join(clientStream pb.ChitChat_JoinServer) error {
	s.mu.Lock()

	// Register new client
	s.lamportTime++
	s.clientStreams = append(s.clientStreams, clientStream)
	clientId := len(s.clientStreams)

	s.mu.Unlock()

	// Send join message
	connectMsg := fmt.Sprintf("Participant #%d joined the ChitChat at logical time %d.", clientId, s.lamportTime)
	log.Println(connectMsg)
	s.SendToClients(connectMsg)

	// Listen to the client while it remains alive
	s.ListenToClient(clientStream, clientId)

	return nil
}

// Continually receive messages from a given client
func (s *Server) ListenToClient(clientStream pb.ChitChat_JoinServer, clientId int) error {
	for {
		// Wait to receive message from client
		in, err := clientStream.Recv()
		s.IncrementLamportTimestamp()

		// If there's an error it's because the client has disconnected (fx. with Ctrl+C)
		if err != nil {
			s.DisconnectClient(clientStream, clientId)
			return err
		}

		// Print and send the client's message
		message := fmt.Sprintf("Client #%d [T:%d]> %s", clientId, s.lamportTime, in.Message)
		log.Println(message)
		s.SendToClients(message)
	}
}

// Disconnect a given client from the server
func (s *Server) DisconnectClient(clientStream pb.ChitChat_JoinServer, clientId int) {
	// Send leave message
	disconnectMsg := fmt.Sprintf("Participant #%d disconnected at logical time %d.", clientId, s.lamportTime)
	log.Println(disconnectMsg)
	s.SendToClients(disconnectMsg)

	// Remove client from slice of clients
	s.clientStreams = RemoveClientFromClientsSlice(s.clientStreams, clientStream)

	// Shut down server if zero clients are left
	if len(s.clientStreams) == 0 {
		log.Println("No clients left: server shutting down...")
		os.Exit(0)
	}
}

// Send a message to all clients listening to the server
func (s *Server) SendToClients(message string) {
	serverMsg := &pb.ServerMessage{
		Message: message,
	}

	for _, c := range s.clientStreams {
		c.Send(serverMsg)
	}
}

// Increment the server's lamport timestamp
func (s *Server) IncrementLamportTimestamp() {
	s.mu.Lock()
	s.lamportTime++
	s.mu.Unlock()
}

func RemoveClientFromClientsSlice(
	clientsSlice []pb.ChitChat_JoinServer,
	client pb.ChitChat_JoinServer) []pb.ChitChat_JoinServer {

	for index, c := range clientsSlice {
		if c == client {
			return append(clientsSlice[:index], clientsSlice[index+1:]...)
		}
	}

	return clientsSlice
}

func main() {
	s := &Server{
		address:       "localhost:50051",
		clientStreams: []pb.ChitChat_JoinServer{},
		lamportTime:   0,
	}

	s.StartServer()
}
