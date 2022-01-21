package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/newtonmunene99/grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	calculatorpb.CalculatorServiceServer
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Add function was invoked with %v\n", req)

	a := req.GetA()
	b := req.GetB()

	sum := a + b

	res := &calculatorpb.SumResponse{
		Sum: sum,
	}

	return res, nil

}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition function was invoked with %v\n", req)

	divisor := int32(2)
	num := req.GetNumber()

	for num > 1 {
		if num%divisor == 0 {
			num = num / divisor
			res := &calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			}
			stream.Send(res)
		} else {
			divisor++
		}
	}

	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {

	fmt.Printf("ComputeAverage function was invoked with a streaming request\n")

	numbers := []int32{}

	for {
		req, err := stream.Recv()

		if err == io.EOF {

			total := int32(0)
			for _, number := range numbers {
				total += number
			}

			average := float64(total) / float64(len(numbers))

			res := &calculatorpb.ComputeAverageResponse{
				Average: average,
			}

			return stream.SendAndClose(res)

		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		number := req.GetNumber()

		numbers = append(numbers, number)
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {

	fmt.Printf("ComputeAverage function was invoked with a streaming request\n")

	max := int32(0)

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		number := msg.GetNumber()

		if number > max {
			max = number

			res := &calculatorpb.FindMaximumResponse{
				Maximum: max,
			}

			sendErr := stream.SendMsg(res)

			if sendErr != nil {
				log.Fatalf("Error while sending data to client: %v", sendErr)
				return sendErr
			}
		}

	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("SquareRoot function was invoked with %v\n", req)

	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v\n", number),
		)
	}

	res := &calculatorpb.SquareRootResponse{
		SquareRoot: math.Sqrt(float64(number)),
	}

	return res, nil

}

func main() {

	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
