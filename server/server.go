package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "module/proto"
)

type Server struct {
	pb.UnimplementedChitChatServer

	mu            sync.Mutex
	address       string
	clientStreams []pb.ChitChat_JoinServer
	lamportTime   int32
}

// Starts a ChitChat server
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
	fmt.Printf("ChitChat Server now listening on %s\n", s.address)

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
	clientId := int32(len(s.clientStreams))

	s.mu.Unlock()

	// Send join message
	log.Printf("Participant #%d joined the ChitChat at logical time %d.", clientId, s.lamportTime)
	fmt.Printf("Participant #%d joined the ChitChat at logical time %d.\n", clientId, s.lamportTime)
	s.SendToClients(&pb.ServerMessage{
		ClientId: clientId,
		Message:  "Participant joined the ChitChat.",
		Lamport:  s.lamportTime,
	})

	// Listen to the client while it remains alive
	s.ListenToClient(clientStream, clientId)

	return nil
}

// Continually receive messages from a given client
func (s *Server) ListenToClient(clientStream pb.ChitChat_JoinServer, clientId int32) error {
	for {
		// Wait to receive message from client
		in, err := clientStream.Recv()

		// Increment Lamport timestamp to whichever is higher of the server's / client's
		s.mu.Lock()
		s.lamportTime = max(s.lamportTime, in.GetLamport()) + 1
		s.mu.Unlock()

		// If there's an error it's because the client has disconnected
		if err != nil {
			s.DisconnectClient(clientStream, clientId)
			return err
		}

		// Print and send the client's message
		log.Printf("Client #%d [T=%d]> %s", clientId, s.lamportTime, in.GetMessage())
		fmt.Printf("Client #%d [T=%d]> %s\n", clientId, s.lamportTime, in.GetMessage())
		s.SendToClients(&pb.ServerMessage{
			ClientId: clientId,
			Message:  in.GetMessage(),
			Lamport:  s.lamportTime,
		})
	}
}

// Disconnect a given client from the server
func (s *Server) DisconnectClient(clientStream pb.ChitChat_JoinServer, clientId int32) {
	// Send leave message (to the server and clients)
	log.Printf("Participant #%d disconnected at logical time %d.", clientId, s.lamportTime)
	fmt.Printf("Participant #%d disconnected at logical time %d.\n", clientId, s.lamportTime)
	s.SendToClients(&pb.ServerMessage{
		ClientId: clientId,
		Message:  "Participant left the ChitChat.",
		Lamport:  s.lamportTime,
	})

	// Remove client from slice of clients and increase Lamport timestamp
	s.mu.Lock()
	s.clientStreams = RemoveClientFromClientsSlice(s.clientStreams, clientStream)
	s.lamportTime++
	s.mu.Unlock()

	// Shut down server if zero clients are left
	if len(s.clientStreams) == 0 {
		log.Println("No clients left: server shutting down...")
		fmt.Println("No clients left: server shutting down...")
		os.Exit(0)
	}
}

// Send a message to all clients listening to the server
func (s *Server) SendToClients(serverMsg *pb.ServerMessage) {
	//serverMsg := &pb.ServerMessage{
	//	Message: message,
	//	Lamport: s.lamportTime,
	//}

	for _, c := range s.clientStreams {
		c.Send(serverMsg)
	}

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

// Creates a log file named after the current time
func CreateLogFile() string {
	year := fmt.Sprint(time.Now().Year())
	month := fmt.Sprint(int(time.Now().Month()))
	day := fmt.Sprint(time.Now().Day())
	hour := fmt.Sprint(time.Now().Hour())
	minute := fmt.Sprint(time.Now().Minute())
	second := fmt.Sprint(time.Now().Second())

	return "server/logs/log-" + year + month + day + "-" + hour + minute + second + ".txt"
}

func main() {
	s := &Server{
		address:       "localhost:50051",
		clientStreams: []pb.ChitChat_JoinServer{},
		lamportTime:   0,
	}

	// Create log file
	fileName := CreateLogFile()

	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	s.StartServer()
}
