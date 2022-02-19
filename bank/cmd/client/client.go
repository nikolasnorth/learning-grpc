package main

import (
	"bufio"
	"context"
	bank "github.com/nikolasnorth/bank/proto"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	portFilename  = "port"
	inputFilename = "input"

	deposit  = "deposit"
	withdraw = "withdraw"
	interest = "interest"
)

func main() {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer inputFile.Close()

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

	// Open gRPC connection
	conn, err := grpc.Dial("localhost:"+port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := bank.NewBankClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Read input file line by line
	scanner = bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.Split(line, " ")
		if len(splitLine) != 3 {
			log.Fatalln("invalid input: expected [operation] [num1] [num2]")
		}

		// Parse current line
		op, accountNumberStr, amountStr := splitLine[0], splitLine[1], splitLine[2]
		accountNumber, err := strconv.ParseInt(accountNumberStr, 10, 64)
		if err != nil {
			log.Fatalf("could not convert string to int64: %v", err)
		}
		amount, err := strconv.ParseFloat(amountStr, 32)
		if err != nil {
			log.Fatalf("could not convert string to float: %v", err)
		}

		// Perform RPC
		req := &bank.Request{AccountNumber: accountNumber, Amount: float32(amount)}
		switch op {
		case deposit:
			_, err := client.Deposit(ctx, req)
			if err != nil {
				log.Fatalf("cannot deposit: %v", err)
			}
		case withdraw:
			_, err := client.Withdraw(ctx, req)
			if err != nil {
				log.Fatalf("cannot withdraw: %v", err)
			}
		case interest:
			_, err := client.AddInterest(ctx, req)
			if err != nil {
				log.Fatalf("cannot add interest: %v", err)
			}
		default:
			log.Fatalf("invalid operation: %s\n", op)
		}
	}
}
