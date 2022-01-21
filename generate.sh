#!/bin/bash

protoc --go_out=. --go-grpc_out=. greet/greetpb/greet.proto

protoc --go_out=. --go-grpc_out=. calculator/calculatorpb/calculator.proto

echo "Generated"