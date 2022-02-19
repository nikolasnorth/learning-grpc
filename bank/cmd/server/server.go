package main

import (
	"bufio"
	"context"
	"encoding/json"
	bank "github.com/nikolasnorth/bank/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	portFilename   = "port"
	inputFilename  = "accounts.json"
	outputFilename = "updatedAccounts.json"
)

type Account struct {
	Name      *string  `json:"Name"`
	AccountId *int64   `json:"AccountID"`
	Balance   *float32 `json:"Balance"`
}

var accounts []Account

type BankServer struct {
	bank.UnimplementedBankServer
}

func (b *BankServer) Deposit(ctx context.Context, req *bank.Request) (*bank.Response, error) {
	for _, account := range accounts {
		if *account.AccountId == req.GetAccountNumber() {
			*account.Balance += req.GetAmount()
		}
	}
	return &bank.Response{}, nil
}

func (b *BankServer) Withdraw(ctx context.Context, req *bank.Request) (*bank.Response, error) {
	for _, account := range accounts {
		if *account.AccountId == req.GetAccountNumber() {
			*account.Balance -= req.GetAmount()
		}
	}
	return &bank.Response{}, nil
}

func (b *BankServer) AddInterest(ctx context.Context, req *bank.Request) (*bank.Response, error) {
	for _, account := range accounts {
		if *account.AccountId == req.GetAccountNumber() {
			*account.Balance += *account.Balance * (req.Amount / 100)
		}
	}
	return &bank.Response{}, nil
}

func main() {
	portFile, err := os.Open(portFilename)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer portFile.Close()

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatalf("cannot open file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("could not create file: %v", err)
	}
	defer outputFile.Close()

	// Unmarshal accounts json data into slice of accounts
	bytes, err := io.ReadAll(inputFile)
	if err != nil {
		log.Fatalf("cannot read from file: %v", err)
	}

	err = json.Unmarshal(bytes, &accounts)
	if err != nil {
		log.Fatalf("cannot unmarshal: %v", err)
	}

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

	// Register Bank service with server
	bank.RegisterBankServer(server, &BankServer{})
	log.Printf("server listening at %v\n", listener.Addr())
	log.Printf("press ctrl-C to shutdown server and generate %s\n", outputFilename)

	// Handle serialization and deserialization
	reflection.Register(server)

	// Listen for SIGINT or error
	sigChan := make(chan os.Signal)
	errChan := make(chan error)

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start server
	go func() {
		err = server.Serve(listener)
		if err != nil {
			errChan <- err
		}
	}()

	// Write updated accounts data to output file upon detecting an OS signal or server error
	defer func() {
		server.GracefulStop()

		bytes, err := json.Marshal(&accounts)
		if err != nil {
			log.Printf("cannot marshal accounts: %v", err)
		}

		_, err = outputFile.Write(bytes)
		if err != nil {
			log.Printf("cannot write to output file: %v", err)
		} else {
			log.Printf("updated accounts written to %s\n", outputFilename)
		}
	}()

	// Block until detecting OS signal or server error
	select {
	case err := <-errChan:
		log.Fatalf("failed to serve: %v", err)
	case <-sigChan:
	}
}
