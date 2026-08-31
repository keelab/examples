// Command sse-stream demonstrates a bounded Server-Sent Events route built
// with the same router and lifecycle contracts as generated streaming code.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/health"
	"github.com/keelab/keelith/metadata"
	"github.com/keelab/keelith/operation"
	transporthttp "github.com/keelab/keelith/transport/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readiness := health.NewRegistry()
	streamServer, err := newStreamServer(readiness)
	if err != nil {
		panic(fmt.Errorf("build stream server: %w", err))
	}
	app, err := keelith.New(
		keelith.WithName("sse-stream"),
		keelith.WithHealth(readiness),
		keelith.WithServer(streamServer),
	)
	if err != nil {
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}

func newStreamServer(readiness *health.Registry) (*transporthttp.Server, error) {
	policy, err := metadata.NewPolicy(nil)
	if err != nil {
		return nil, fmt.Errorf("build metadata policy: %w", err)
	}
	router, err := transporthttp.NewRouter(transporthttp.WithMetadataPolicy(policy))
	if err != nil {
		return nil, fmt.Errorf("build HTTP router: %w", err)
	}
	encoder, err := transporthttp.NewSSEEncoder(transporthttp.SSEConfig{
		DisableHeartbeat: true,
	})
	if err != nil {
		return nil, fmt.Errorf("build SSE encoder: %w", err)
	}
	target, err := operation.New(
		"http",
		"examples.v1.StreamService",
		"Events",
		operation.KindServerStream,
	)
	if err != nil {
		return nil, fmt.Errorf("build stream operation: %w", err)
	}
	if err := router.Handle(
		http.MethodGet,
		"/events",
		target,
		transporthttp.DecodeSSERequest,
		streamEvents,
		encoder,
		transporthttp.WithStreaming(),
	); err != nil {
		return nil, fmt.Errorf("register stream route: %w", err)
	}
	server, err := transporthttp.NewServer(
		router,
		transporthttp.WithAddress("127.0.0.1:18092"),
		transporthttp.WithName("sse-stream-http"),
		transporthttp.WithHealth(readiness),
	)
	if err != nil {
		return nil, fmt.Errorf("build HTTP server: %w", err)
	}
	return server, nil
}

func streamEvents(ctx context.Context, request any) (any, error) {
	sseRequest, ok := request.(transporthttp.SSERequest)
	if !ok {
		return nil, fmt.Errorf("unexpected SSE request type %T", request)
	}
	lastID, err := strconv.Atoi(sseRequest.LastEventID())
	if err != nil && sseRequest.LastEventID() != "" {
		return nil, fmt.Errorf("parse Last-Event-ID: %w", err)
	}
	events := make(chan transporthttp.SSEEvent, 1)
	failures := make(chan error, 1)
	go func() {
		defer close(events)
		defer close(failures)
		for id := lastID + 1; id <= 3; id++ {
			event, eventErr := transporthttp.NewSSEJSONEvent(
				strconv.Itoa(id),
				"tick",
				map[string]int{"sequence": id},
				0,
			)
			if eventErr != nil {
				failures <- eventErr
				return
			}
			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}()
	return transporthttp.NewServerSentEvents(events, failures)
}
