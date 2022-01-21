package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/newtonmunene99/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doBiDirectionalStreaming(c)

	doErrorUnary(c)

}

func doUnary(c calculatorpb.CalculatorServiceClient) {

	req := &calculatorpb.SumRequest{
		A: 3,
		B: 10,
	}

	sum, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling Add RPC: %v", err)
	}

	log.Printf("Response from Add: %v", sum.Sum)

}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 120,
	}

	resStream, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling PrimeNumberDecomposition RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}

		log.Printf("Response from PrimeNumberDecomposition: %v", msg.GetPrimeFactor())

	}

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {

	stream, err := c.ComputeAverage(context.Background())

	if err != nil {
		log.Fatalf("Error while calling ComputeAverage: %v", err)
	}

	requests := []*calculatorpb.ComputeAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}

	for _, request := range requests {

		fmt.Printf("Sending req: %v\n", request)

		stream.SendMsg(request)

		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while receiving response from ComputeAverage: %v\n", err)
	}

	fmt.Printf("ComputeAverage Response: %v\n", res)
}

func doBiDirectionalStreaming(c calculatorpb.CalculatorServiceClient) {
	stream, err := c.FindMaximum(context.Background())

	if err != nil {
		log.Fatalf("Error while calling FindMaximum: %v", err)
	}

	requests := []*calculatorpb.FindMaximumRequest{
		{
			Number: 1,
		},
		{
			Number: 5,
		},
		{
			Number: 3,
		},
		{
			Number: 6,
		},
		{
			Number: 2,
		},
		{
			Number: 20,
		},
	}

	waitc := make(chan struct{})

	go func() {

		for _, request := range requests {
			fmt.Printf("Sending req: %v\n", request)

			err := stream.Send(request)

			if err != nil {
				log.Fatalf("Error while sending request: %v", err)
			}

			time.Sleep(1000 * time.Millisecond)
		}

		stream.CloseSend()
	}()

	go func() {
		for {
			msg, err := stream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Error while receiving response: %v", err)

				break
			}

			fmt.Printf("Received: %v\n", msg.GetMaximum())

		}

		close(waitc)
	}()

	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a SquareRoot Unary RPC")

	req := &calculatorpb.SquareRootRequest{
		Number: 10,
	}

	res, err := c.SquareRoot(context.Background(), req)

	if err != nil {
		resErr, ok := status.FromError(err)

		if ok {

			fmt.Printf("Error message: %v\n", resErr.Message())

			fmt.Printf("Error code: %v\n", resErr.Code())

			if resErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
			}

		} else {
			log.Fatalf("Error while calling SquareRoot: %v", err)
		}
	}

	fmt.Printf("Response from SquareRoot: %v\n", res.GetSquareRoot())
}
