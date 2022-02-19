package main

import (
	"bufio"
	"context"
	calculator "github.com/nikolasnorth/calculator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

const portFilename = "port"

type CalculatorServer struct {
	calculator.UnimplementedCalculatorServer
}

func (c *CalculatorServer) Add(ctx context.Context, req *calculator.IntRequest) (*calculator.IntResponse, error) {
	result := req.GetA() + req.GetB()
	return &calculator.IntResponse{Result: result}, nil
}

func (c *CalculatorServer) Sub(ctx context.Context, req *calculator.IntRequest) (*calculator.IntResponse, error) {
	result := req.GetA() - req.GetB()
	return &calculator.IntResponse{Result: result}, nil
}

func (c *CalculatorServer) Mult(ctx context.Context, req *calculator.IntRequest) (*calculator.IntResponse, error) {
	result := req.GetA() * req.GetB()
	return &calculator.IntResponse{Result: result}, nil
}

func (c *CalculatorServer) Div(ctx context.Context, req *calculator.IntRequest) (*calculator.FloatResponse, error) {
	result := float32(req.GetA()) / float32(req.GetB())
	return &calculator.FloatResponse{Result: result}, nil
}

func main() {
	portFile, err := os.Open(portFilename)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer portFile.Close()

	// Read port number from file
	scanner := bufio.NewScanner(portFile)
	port := ""
	if scanner.Scan() {
		port = scanner.Text()
	}
	err = scanner.Err()
	if err != nil {
		log.Fatalf("scanner error: %v", err)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create new gRPC server
	server := grpc.NewServer()

	// Register Calculator service with server
	calculator.RegisterCalculatorServer(server, &CalculatorServer{})
	log.Printf("server listening at %v", listener.Addr())

	// Handle serialization and deserialization
	reflection.Register(server)

	err = server.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
