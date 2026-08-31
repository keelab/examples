package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/keelith"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := keelith.New(
		keelith.WithName("hello-http"),
		keelith.WithHTTP(":8080"),
		keelith.WithRoute(http.MethodGet, "/ping", func(ctx context.Context, req *http.Request) (any, error) {
			return map[string]string{
				"message": "pong",
			}, nil
		}),
	)
	if err != nil {
		panic(err)
	}

	if err = app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
