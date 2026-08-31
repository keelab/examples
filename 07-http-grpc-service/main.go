// Command http-grpc-service demonstrates one generated service binding
// mounted on both the HTTP and gRPC listeners.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	pingv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
	"github.com/keelab/examples/07-http-grpc-service/internal/ping"
	"github.com/keelab/keelith"
	"github.com/keelab/keelith/middleware"
	"github.com/keelab/keelith/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bundle, err := middleware.NewBundle(middleware.Entry{
		Name: "service-log",
		Middleware: func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, request any) (any, error) {
				log.Print("service middleware: ping.v1.PingService/Ping")
				return next(ctx, request)
			}
		},
	})
	if err != nil {
		panic(err)
	}

	profile, err := service.NewProfile(
		"ping",
		pingv1.BindPingService(
			ping.New(),
			service.WithHTTPBundle(bundle),
			service.WithGRPCBundle(bundle),
		),
	)
	if err != nil {
		panic(err)
	}

	app, err := keelith.New(
		keelith.WithName("http-grpc-service"),
		keelith.WithHTTP(":8086"),
		keelith.WithGRPC(":8087"),
		keelith.WithProfile(profile),
	)
	if err != nil {
		panic(err)
	}

	if err = app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
