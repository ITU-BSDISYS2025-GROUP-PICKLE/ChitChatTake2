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

// This function runs for every client that joins the server
func (s *Server) Join(clientStream pb.ChitChat_JoinServer) error {
	s.mu.Lock()

	// Register new client
	s.clientStreams = append(s.clientStreams, clientStream)
	clientId := len(s.clientStreams)

	s.mu.Unlock()

	// Send join message
	connectMsg := fmt.Sprintf("Participant #%d joined the ChitChat.", clientId)
	log.Println(connectMsg)
	SendToClients(s, connectMsg)

	// Listen to the client while it remains alive
	ListenToClient(clientStream, clientId, s)

	return nil
}

// Continually receive messages from a given client
func ListenToClient(clientStream pb.ChitChat_JoinServer, clientId int, s *Server) error {
	for {
		// Wait to receive message from client
		in, err := clientStream.Recv()

		// If there's an error it's because the client has disconnected (fx. with Ctrl+C)
		if err != nil {
			DisconnectClient(clientStream, clientId, s)
			return err
		}

		// Print and send the client's message
		message := fmt.Sprintf("Client #%d [local logical time: %d]> %s", clientId, in.Lamport, in.Message)
		log.Println(message)
		SendToClients(s, message)
	}
}

// Disconnect a given client from the server
func DisconnectClient(clientStream pb.ChitChat_JoinServer, clientId int, s *Server) {
	// Send leave message
	disconnectMsg := fmt.Sprintf("Participant #%d disconnected.", clientId)
	log.Println(disconnectMsg)
	SendToClients(s, disconnectMsg)

	// Remove client from slice of clients
	s.clientStreams = RemoveClientFromClientsSlice(s.clientStreams, clientStream)

	// Shut down server if zero clients are left
	if len(s.clientStreams) == 0 {
		log.Println("No clients left: shutting down server...")
		os.Exit(0)
	}
}

// Send a message to all clients listening to the server
func SendToClients(s *Server, message string) {
	serverMsg := &pb.ServerMessage{
		Message: message,
	}

	for _, c := range s.clientStreams {
		c.Send(serverMsg)
	}
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
	}

	s.StartServer()
}
