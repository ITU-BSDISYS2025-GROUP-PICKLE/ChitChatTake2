package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "module/proto"
)

// Dial the server, create a client
func CreateClient() pb.ChitChatClient {
	// Create connection
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect to server failed: %v", err)
	}

	// Create and return client
	return pb.NewChitChatClient(conn)
}

// Continually receive messages from the server
func ListenToServer(serverStream grpc.BidiStreamingClient[pb.ClientMessage, pb.ServerMessage]) {
	for {
		in, err := serverStream.Recv()
		if err != nil {
			println("Mi bombo")
			return
		}

		// Print received messages
		println(in.Message)
	}
}

// Read user input. Considers message-length and includes exit-words.
func ReadInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	input := scanner.Text()

	if len(input) > 128 {
		println("Error: Message too long")
		return ""
	}

	// Disconnect if input matches a word in the exitWords slice
	exitWords := []string{".exit", "/exit", "--exit", ".e", "/e", "-e"}
	for _, exitWord := range exitWords {
		if input == exitWord {
			os.Exit(0)
		}
	}

	return input
}

// Send a message to the server
func SendToServer(serverStream grpc.BidiStreamingClient[pb.ClientMessage, pb.ServerMessage], message string) {
	if err := serverStream.Send(&pb.ClientMessage{Message: message}); err != nil {
		log.Fatalf("serverStream.Send() failed: %v", err)
	}
}

func main() {
	// Create client
	client := CreateClient()

	// Create context with 10-minute timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Get server-stream via RPC
	serverStream, err := client.Join(ctx)
	if err != nil {
		log.Fatalf("client.Join failed: %v", err)
	}

	// Continually receive messages from the server
	go ListenToServer(serverStream)

	// Continually read and send messages to the server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		message := ReadInput(scanner)
		time.Sleep(100 * time.Millisecond)
		// ^^Fixes an annoying behaviour where Client terminal prints
		// 4x "Error: Message cannot be empty" if Ctrl+C is used to disconnect

		if strings.Trim(message, " ") == "" {
			println("Error: Message cannot be empty")
			continue
		}

		SendToServer(serverStream, message)
	}
}
