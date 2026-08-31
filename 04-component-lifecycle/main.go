package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/keelith"
	kapp "github.com/keelab/keelith/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	component := kapp.ComponentFunc{
		ComponentName: "in-memory-cache",
		StartFunc: func(context.Context) error {
			log.Print("cache started")
			return nil
		},
		StopFunc: func(context.Context) error {
			log.Print("cache stopped")
			return nil
		},
	}
	app, err := keelith.New(
		keelith.WithName("component-lifecycle"),
		keelith.WithHTTP(":8083"),
		keelith.WithComponent(component),
		keelith.WithRoute(http.MethodGet, "/status", func(context.Context, *http.Request) (any, error) {
			return map[string]string{
				"status": "ready",
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
