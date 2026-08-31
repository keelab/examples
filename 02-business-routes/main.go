package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/keelab/keelith"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := keelith.New(
		keelith.WithName("business-routes"),
		keelith.WithHTTP(":8080"),
		keelith.WithRoute(http.MethodGet, "/healthz", health),
		keelith.WithRoute(http.MethodGet, "/greeting", greeting),
	)
	if err != nil {
		panic(err)
	}

	if err = app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}

func health(ctx context.Context, req *http.Request) (any, error) {
	return map[string]string{
		"status": "up",
	}, nil
}

func greeting(ctx context.Context, req *http.Request) (any, error) {
	name := strings.TrimSpace(req.URL.Query().Get("name"))
	if name == "" {
		name = "World"
	}
	return map[string]string{
		"message": "heelo " + name,
	}, nil
}
