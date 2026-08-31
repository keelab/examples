// Command client calls the generated gRPC client from the parent example.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pingv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	connection, err := grpc.NewClient(
		"127.0.0.1:8087",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			log.Printf("close gRPC connection: %v", err)
		}
	}()

	client := pingv1.NewPingServiceGRPCClient(connection)
	if _, err = client.Ping(ctx, &emptypb.Empty{}); err != nil {
		panic(err)
	}
	fmt.Println("gRPC response: pong")
}
