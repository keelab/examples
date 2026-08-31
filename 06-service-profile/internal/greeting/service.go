package greeting

import (
	"context"

	greetingv1 "github.com/keelab/examples/06-service-profile/gen/greeting/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Service is the business implementation shared by generated HTTP and gRPC bindings.
type Service struct{}

// New constructs the example service without external dependencies.
func New() *Service {
	return &Service{}
}

// GetGreeting returns a greeting payload for the profile binding example.
func (*Service) GetGreeting(context.Context, *emptypb.Empty) (*greetingv1.GetGreetingResponse, error) {
	return &greetingv1.GetGreetingResponse{Message: "hello from a profile binding"}, nil
}

var _ greetingv1.GreetingServiceKeelithServer = (*Service)(nil)
