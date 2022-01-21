package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/newtonmunene99/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello I'm a gRPC client")

	tls := true
	opts := grpc.WithInsecure()

	if tls {

		certFile := "ssl/ca.crt"

		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")

		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
			return
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doBiDirectionalStreaming(c)

	// doUnaryWithDeadline(c, 1*time.Second)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Unary RPC")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Newton",
			LastName:  "Munene",
		},
	}

	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}

	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {

	fmt.Println("Starting to do Server Streaming RPC")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Newton",
			LastName:  "Munene",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}

		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Client Streaming RPC")

	stream, err := c.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("Error while calling LongGreet: %v", err)
	}

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Newton",
				LastName:  "Munene",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Kelvin",
				LastName:  "Murimi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Brian",
				LastName:  "Mwaniki",
			},
		},
	}

	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)

		stream.Send(req)

		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet: %v\n", err)
	}

	fmt.Printf("LongGreet Response: %v\n", res)

}

func doBiDirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Bi-Directional Streaming RPC")

	stream, err := c.GreetEveryone(context.Background())

	if err != nil {
		log.Fatalf("Error while calling GreetEveryone: %v", err)

	}

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Newton",
				LastName:  "Munene",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Kelvin",
				LastName:  "Murimi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Brian",
				LastName:  "Mwaniki",
			},
		},
	}

	waitc := make(chan struct{})

	go func() {
		for _, request := range requests {
			fmt.Printf("Sending message: %v\n", request)

			stream.Send(request)

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
				log.Fatalf("Error while receiving: %v", err)

				break
			}

			fmt.Printf("Received: %v\n", msg.GetResult())
		}
		close(waitc)
	}()

	<-waitc

}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting to do Unary RPC with Deadline")

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Newton",
			LastName:  "Munene",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)

	if err != nil {

		statusErr, ok := status.FromError(err)

		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("Unexpected error: %v\n", statusErr)
			}
		} else {

			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
		}

		return
	}

	log.Printf("Response from GreetWithDeadline: %v", res.Result)
}
