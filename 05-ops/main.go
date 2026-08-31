package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/ops"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := keelith.New(
		keelith.WithName("ops-example"),
		keelith.WithHTTP(":8084"),
		keelith.WithOps(
			ops.WithAddress("127.0.0.1:9090"),
			ops.WithBuildInfo(ops.CurrentBuildInfo()),
		),
		keelith.WithRoute(http.MethodGet, "/ping", func(context.Context, *http.Request) (any, error) {
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
