// Command cache demonstrates a typed read-through cache, request coalescing
// and versioned invalidation with the in-memory backend.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/cache"
	cacheMemory "github.com/keelab/keelith/cache/memory"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var loads atomic.Uint64
	store, err := cache.New(
		cacheMemory.New(),
		cache.JSONCodec[string]{},
		func(context.Context, string) (string, error) {
			sequence := loads.Add(1)
			return "loaded-" + strconv.FormatUint(sequence, 10), nil
		},
		cache.Policy{
			TTL:         time.Minute,
			JitterRatio: 0,
		},
		cache.WithVersioning(),
	)
	if err != nil {
		panic(fmt.Errorf("build cache: %w", err))
	}

	app, err := keelith.New(
		keelith.WithName("cache"),
		keelith.WithHTTP(":8091"),
		keelith.WithRoute(http.MethodGet, "/value", func(ctx context.Context, _ *http.Request) (any, error) {
			value, getErr := store.Get(ctx, "welcome")
			if getErr != nil {
				return nil, fmt.Errorf("get cached value: %w", getErr)
			}
			return map[string]any{
				"value": value,
				"loads": loads.Load(),
			}, nil
		}),
		keelith.WithRoute(http.MethodPost, "/invalidate", func(ctx context.Context, _ *http.Request) (any, error) {
			state, invalidateErr := store.InvalidateVersion(ctx, "welcome", 1)
			if invalidateErr != nil && !errors.Is(invalidateErr, cache.ErrVersioningUnsupported) {
				return nil, fmt.Errorf("invalidate cached value: %w", invalidateErr)
			}
			return map[string]string{"state": fmt.Sprint(state)}, nil
		}),
	)
	if err != nil {
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}
