package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "distributed-printer/proto"
	"google.golang.org/grpc"
)

type printServer struct {
	pb.UnimplementedPrintServiceServer
}

func (s *printServer) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	log.Printf("Received print request: %s", req.Message)

	// Simulate printing
	fmt.Println("PRINTING:", req.Message)

	return &pb.PrintResponse{
		Success: true,
		Result:  fmt.Sprintf("Printed: %s", req.Message),
	}, nil
}

func main() {
	port := "50051"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPrintServiceServer(s, &printServer{})

	log.Printf("Print server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
