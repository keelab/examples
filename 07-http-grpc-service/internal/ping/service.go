package ping

import (
	"context"

	pingv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Service is the business implementation shared by generated HTTP and gRPC bindings.
type Service struct{}

// New constructs the example service without external dependencies.
func New() *Service {
	return &Service{}
}

// Ping returns an empty successful response.
func (*Service) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

var _ pingv1.PingServiceKeelithServer = (*Service)(nil)
