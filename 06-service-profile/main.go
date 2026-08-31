package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/examples/06-service-profile/internal/greeting"
	"github.com/keelab/keelith"
	greetingv1 "github.com/keelab/examples/06-service-profile/gen/greeting/v1"
	"github.com/keelab/keelith/middleware"
	"github.com/keelab/keelith/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bundle, err := middleware.NewBundle(middleware.Entry{
		Name: "request-log",
		Middleware: func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, request any) (any, error) {
				log.Print("request middleware: greeting.v1.GreetingService/GetGreeting")
				return next(ctx, request)
			}
		},
	})
	if err != nil {
		panic(err)
	}

	profile, err := service.NewProfile("greeting", greetingv1.BindGreetingService(greeting.New(), service.WithHTTPBundle(bundle)))
	if err != nil {
		panic(err)
	}

	app, err := keelith.New(
		keelith.WithName("service-profile"),
		keelith.WithHTTP(":8085"),
		keelith.WithProfile(profile),
	)
	if err != nil {
		panic(err)
	}

	if err = app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
