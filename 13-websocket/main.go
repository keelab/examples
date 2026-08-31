// Command websocket demonstrates the lifecycle-owned WebSocket Hub and
// transport-neutral stream middleware.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/health"
	"github.com/keelab/keelith/metadata"
	"github.com/keelab/keelith/middleware"
	"github.com/keelab/keelith/operation"
	transporthttp "github.com/keelab/keelith/transport/http"
	transportws "github.com/keelab/keelith/transport/websocket"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	streamMiddleware, err := middleware.NewStreamBundle(middleware.StreamEntry{
		Name: "stream-log",
		Middleware: func(next middleware.StreamHandler) middleware.StreamHandler {
			return func(ctx context.Context, event middleware.StreamEvent) error {
				log.Printf("websocket stream phase=%s", event.Phase)
				return next(ctx, event)
			}
		},
	})
	if err != nil {
		panic(fmt.Errorf("build stream middleware: %w", err))
	}
	hub, err := transportws.NewHub(transportws.Options{
		Name:           "example-websocket",
		OriginPatterns: []string{"http://localhost:*", "http://127.0.0.1:*"},
		MaxConnections: 32,
	}, streamMiddleware)
	if err != nil {
		panic(fmt.Errorf("build websocket hub: %w", err))
	}
	readiness := health.NewRegistry()
	server, err := newWebSocketServer(hub, readiness)
	if err != nil {
		panic(err)
	}
	app, err := keelith.New(
		keelith.WithName("websocket"),
		keelith.WithHealth(readiness),
		keelith.WithComponent(hub),
		keelith.WithServer(server),
	)
	if err != nil {
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}

func newWebSocketServer(
	hub *transportws.Hub,
	readiness *health.Registry,
) (*transporthttp.Server, error) {
	policy, err := metadata.NewPolicy(nil)
	if err != nil {
		return nil, fmt.Errorf("build metadata policy: %w", err)
	}
	router, err := transporthttp.NewRouter(transporthttp.WithMetadataPolicy(policy))
	if err != nil {
		return nil, fmt.Errorf("build HTTP router: %w", err)
	}
	target, err := operation.New(
		"http",
		"examples.v1.WebSocketService",
		"Connect",
		operation.KindBidiStream,
	)
	if err != nil {
		return nil, fmt.Errorf("build websocket operation: %w", err)
	}
	if err := router.Handle(
		http.MethodGet,
		"/ws",
		target,
		transportws.DecodeRequest,
		websocketSession,
		hub.Encode,
		transporthttp.WithStreaming(),
	); err != nil {
		return nil, fmt.Errorf("register websocket route: %w", err)
	}
	server, err := transporthttp.NewServer(
		router,
		transporthttp.WithAddress("127.0.0.1:8093"),
		transporthttp.WithName("websocket-http"),
		transporthttp.WithHealth(readiness),
	)
	if err != nil {
		return nil, fmt.Errorf("build HTTP server: %w", err)
	}
	return server, nil
}

func websocketSession(_ context.Context, request any) (any, error) {
	websocketRequest, ok := request.(transportws.Request)
	if !ok {
		return nil, fmt.Errorf("unexpected websocket request type %T", request)
	}
	return transportws.NewSession(websocketRequest, func(ctx context.Context, connection *transportws.Connection) error {
		message, err := connection.Read(ctx)
		if err != nil {
			return err
		}
		return connection.Write(ctx, message)
	})
}
