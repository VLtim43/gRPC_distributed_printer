package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	pb "distributed-printer/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Auto-generate unique client ID (using Int31 for shorter IDs)
	clientID := fmt.Sprintf("%d", rand.Int31())

	fmt.Printf("========== CLIENT %s ==========\n", clientID)
	fmt.Println()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPrintServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Connected to print server as CLIENT %s. Type messages to print (Ctrl+C to exit):\n", clientID)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := client.Print(ctx, &pb.PrintRequest{
			Message:  message,
			ClientId: clientID,
		})
		cancel()

		if err != nil {
			log.Printf("Failed to print: %v", err)
			continue
		}

		fmt.Printf("Server: %s\n", resp.Result)
	}
}
