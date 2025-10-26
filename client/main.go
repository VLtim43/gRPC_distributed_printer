package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "distributed-printer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("========== CLIENT ==========")
	fmt.Println()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPrintServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Connected to print server. Type messages to print (Ctrl+C to exit):")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.Print(ctx, &pb.PrintRequest{
			Message: message,
		})
		cancel()

		if err != nil {
			log.Printf("Failed to print: %v", err)
			continue
		}

		fmt.Printf("Server: %s\n", resp.Result)
	}
}
