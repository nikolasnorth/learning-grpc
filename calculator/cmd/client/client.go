package main

import (
	"bufio"
	"context"
	"fmt"
	calculator "github.com/nikolasnorth/calculator/proto"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	portFilename   = "port"
	inputFilename  = "input"
	outputFilename = "output"

	add  string = "add"
	sub  string = "sub"
	mult string = "mult"
	div  string = "div"
)

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

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("could not create file: %v", err)
	}
	defer outputFile.Close()

	// Open gRPC connection
	conn, err := grpc.Dial("localhost:"+port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := calculator.NewCalculatorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Read input file line by line
	scanner = bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.Split(line, " ")
		if len(splitLine) != 3 {
			log.Fatalln("invalid input, expected: [operation] [num1] [num2]")
		}

		// Parse current line
		op, aStr, bStr := splitLine[0], splitLine[1], splitLine[2]
		a, err := strconv.ParseInt(aStr, 10, 64)
		if err != nil {
			log.Fatalf("could not convert string to int64: %v", err)
		}
		b, err := strconv.ParseInt(bStr, 10, 64)
		if err != nil {
			log.Fatalf("could not convert string to int64: %v", err)
		}

		// Perform RPC and write response to output file
		req := &calculator.IntRequest{A: a, B: b}
		switch op {
		case add:
			res, err := client.Add(ctx, req)
			if err != nil {
				log.Fatalf("could not add %d and %d: %v", req.GetA(), req.GetB(), err)
			}

			_, err = outputFile.WriteString(strconv.FormatInt(res.GetResult(), 10) + "\n")
			if err != nil {
				log.Fatalf("could not write result to output file: %v", err)
			}
		case sub:
			res, err := client.Sub(ctx, req)
			if err != nil {
				log.Fatalf("could not subtract %d and %d: %v", req.GetA(), req.GetB(), err)
			}

			_, err = outputFile.WriteString(strconv.FormatInt(res.GetResult(), 10) + "\n")
			if err != nil {
				log.Fatalf("could not write result to output file: %v", err)
			}
		case mult:
			res, err := client.Mult(ctx, req)
			if err != nil {
				log.Fatalf("could not multiply %d and %d: %v", req.GetA(), req.GetB(), err)
			}

			_, err = outputFile.WriteString(strconv.FormatInt(res.GetResult(), 10) + "\n")
			if err != nil {
				log.Fatalf("could not write result to output file: %v", err)
			}
		case div:
			res, err := client.Div(ctx, req)
			if err != nil {
				log.Fatalf("could not divide %d and %d: %v", req.GetA(), req.GetB(), err)
			}

			_, err = outputFile.WriteString(fmt.Sprintf("%.2f", res.GetResult()) + "\n")
			if err != nil {
				log.Fatalf("could not write result to output file: %v", err)
			}
		default:
			log.Fatalf("invalid operator: %s\n", op)
		}
	}
}
