package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "distributed-printer/proto"

	"google.golang.org/grpc"
)

type printServer struct {
	pb.UnimplementedPrintServiceServer
}

func (s *printServer) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	timestamp := time.Now().Unix()

	printOutput := fmt.Sprintf("[TS: %d] CLIENT %s: %s", timestamp, req.ClientId, req.Message)
	fmt.Println(printOutput)

	return &pb.PrintResponse{
		Success: true,
		Result:  fmt.Sprintf("Print job completed for client %s", req.ClientId),
	}, nil
}

func main() {
	fmt.Println("========== SERVER ==========")
	fmt.Println()

	port := "50051"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPrintServiceServer(s, &printServer{})

	// shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down server gracefully...")
		s.GracefulStop()
		os.Exit(0)
	}()

	log.Printf("Print server listening on port %s", port)
	log.Println("Press Ctrl+C to stop the server")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
