package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/config"
)

var configuredMessage atomic.Value

type binding struct{}

func (binding) Name() string { return "example" }

func (binding) Validate(snapshot config.Snapshot) error {
	value, ok := snapshot.Lookup("message")
	if !ok {
		return fmt.Errorf("message is required")
	}
	if _, ok := value.(string); !ok {
		return fmt.Errorf("message must be a string")
	}
	return nil
}

func (binding) Apply(_ context.Context, snapshot config.Snapshot) error {
	value, ok := snapshot.Lookup("message")
	if !ok {
		return fmt.Errorf("message disappeared before apply")
	}
	message, ok := value.(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}
	configuredMessage.Store(message)
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := keelith.New(
		keelith.WithName("file-config"),
		keelith.WithHTTP(":8082"),
		keelith.WithConfigFile(
			"03-file-config/configs/dev.yaml",
			keelith.WithConfigEnvPrefix("KEELITH_CONFIG"),
			keelith.WithConfigKnownFields("message"),
			keelith.WithConfigBindings(binding{}),
			keelith.WithConfigStrict(),
			keelith.WithConfigPollInterval(time.Second),
		),
		keelith.WithRoute(http.MethodGet, "/message", message),
	)
	if err != nil {
		panic(err)
	}

	if err = app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}

func message(context.Context, *http.Request) (any, error) {
	value, ok := configuredMessage.Load().(string)
	if !ok {
		return nil, fmt.Errorf("configuration has not loaded")
	}
	return map[string]string{"message": value}, nil
}
