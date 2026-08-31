// Command di-graph demonstrates typed providers, dependency construction and
// graph cleanup attached to a Keelith Application.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/di"
)

type greetingConfig struct {
	Prefix string
}

type greeter struct {
	config greetingConfig
}

func provideConfig() greetingConfig {
	return greetingConfig{Prefix: "hello"}
}

func provideGreeter(config greetingConfig) *greeter {
	return &greeter{config: config}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	module, err := di.NewModule(
		"example",
		di.Provide(provideConfig),
		di.Provide(provideGreeter),
	)
	if err != nil {
		panic(fmt.Errorf("build DI module: %w", err))
	}
	graph, service, err := di.Build[*greeter](ctx, module)
	if err != nil {
		panic(fmt.Errorf("build DI graph: %w", err))
	}

	app, err := keelith.New(
		keelith.WithName("di-graph"),
		keelith.WithHTTP(":8094"),
		keelith.WithGraph(graph),
		keelith.WithRoute(http.MethodGet, "/greeting", func(context.Context, *http.Request) (any, error) {
			return map[string]string{"message": service.config.Prefix + " from DI"}, nil
		}),
	)
	if err != nil {
		closeErr := graph.Close(context.Background())
		if closeErr != nil {
			panic(errors.Join(
				fmt.Errorf("build application: %w", err),
				fmt.Errorf("close DI graph: %w", closeErr),
			))
		}
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}
