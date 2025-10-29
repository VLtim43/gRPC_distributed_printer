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

	"distributed-printer/clock"
	pb "distributed-printer/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Auto-generate unique client ID
	clientID := fmt.Sprintf("%d", rand.Int31())

	lamportClock := clock.New()

	fmt.Printf("========== CLIENT %s ==========\n", clientID)
	fmt.Println()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPrintServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	log.Printf("Connected to print server as CLIENT %s", clientID)
	log.Println("Type messages to print (Ctrl+C to exit)")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		// Increment Lamport clock before sending message
		timestamp := lamportClock.Increment()
		fmt.Printf("\n[Lamport Clock: %d] Sending print request...\n", timestamp)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := client.Print(ctx, &pb.PrintRequest{
			Message:  message,
			ClientId: clientID,
		})
		cancel()

		if err != nil {
			log.Printf("Failed to print: %v", err)
			continue
		}

		fmt.Printf("[Lamport Clock: %d] Server: %s\n", lamportClock.Get(), resp.Result)
	}
}
