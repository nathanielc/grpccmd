package main

import (
	"log"
	"net"

	"github.com/nathanielc/grpccmd/example/internal/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct{}

// GetNumber returns a number
func (s *server) GetNumber(ctx context.Context, in *pb.Empty) (*pb.Number, error) {
	log.Println("GetNumber")
	return &pb.Number{Value: 42}, nil
}

// Echo returns the data it is given
func (s *server) Echo(ctx context.Context, in *pb.EchoData) (*pb.EchoData, error) {
	log.Println("Echo")
	return in, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterExampleServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("Listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
