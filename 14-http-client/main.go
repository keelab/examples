// Command http-client demonstrates a typed HTTP client call with outbound
// middleware and allowlisted metadata propagation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/keelab/keelith/metadata"
	"github.com/keelab/keelith/middleware"
	"github.com/keelab/keelith/operation"
	transporthttp "github.com/keelab/keelith/transport/http"
)

type pingResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := json.NewEncoder(writer).Encode(pingResponse{
			Message:   "pong",
			RequestID: request.Header.Get("X-Request-Id"),
		}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	bundle, err := middleware.NewBundle(middleware.Entry{
		Name: "client-log",
		Middleware: func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, request any) (any, error) {
				target, ok := operation.FromContext(ctx)
				if !ok {
					return nil, fmt.Errorf("outbound operation is missing")
				}
				fmt.Printf("outbound operation: %s\n", target)
				return next(ctx, request)
			}
		},
	})
	if err != nil {
		panic(fmt.Errorf("build client middleware: %w", err))
	}
	policy, err := metadata.NewPolicy([]string{"x-request-id"})
	if err != nil {
		panic(fmt.Errorf("build metadata policy: %w", err))
	}
	client, err := transporthttp.NewClient(
		server.Client(),
		transporthttp.WithClientMetadataPolicy(policy),
		transporthttp.WithClientMiddleware(bundle),
	)
	if err != nil {
		panic(fmt.Errorf("build HTTP client: %w", err))
	}
	request, err := http.NewRequest(http.MethodGet, server.URL+"/v1/ping", nil)
	if err != nil {
		panic(fmt.Errorf("build request: %w", err))
	}
	target, err := operation.New("http", "examples.v1.RemoteService", "Ping", operation.KindUnary)
	if err != nil {
		panic(fmt.Errorf("build operation: %w", err))
	}
	outbound, err := metadata.New(map[string][]string{
		"x-request-id": {"example-request"},
	})
	if err != nil {
		panic(fmt.Errorf("build outbound metadata: %w", err))
	}
	response, err := transporthttp.Invoke(
		metadata.WithOutbound(context.Background(), outbound),
		client,
		target,
		transporthttp.ClientCall[pingResponse]{
			Request: request,
			Decode: func(_ context.Context, response *http.Response) (pingResponse, error) {
				var value pingResponse
				if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
					return pingResponse{}, err
				}
				return value, nil
			},
		},
	)
	if err != nil {
		panic(fmt.Errorf("invoke ping: %w", err))
	}
	fmt.Printf("response: message=%s request_id=%s\n", response.Message, response.RequestID)
}
