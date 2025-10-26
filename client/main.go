package main

import (
	"context"
	"log"
	"time"

	pb "distributed-printer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPrintServiceClient(conn)

	// Send a print request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.Print(ctx, &pb.PrintRequest{
		Message: "Hello from client!",
	})
	if err != nil {
		log.Fatalf("Failed to print: %v", err)
	}

	log.Printf("Server response: success=%v, result=%s", resp.Success, resp.Result)
}
