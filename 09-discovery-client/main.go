// Command discovery-client demonstrates in-process registration, discovery,
// selection and the application-owned client Router lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/client"
	"github.com/keelab/keelith/operation"
	"github.com/keelab/keelith/registry"
	registryMemory "github.com/keelab/keelith/registry/memory"
	"github.com/keelab/keelith/selector"
)

const discoveryService = "keelith.quickstart"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	registryBackend := registryMemory.New()
	for _, instance := range []struct {
		id       string
		endpoint string
		zone     string
	}{
		{id: "node-a", endpoint: "http://127.0.0.1:18081", zone: "zone-a"},
		{id: "node-c", endpoint: "http://127.0.0.1:18083", zone: "zone-a"},
		{id: "node-b", endpoint: "http://127.0.0.1:18082", zone: "zone-b"},
	} {
		value, err := registry.NewInstance(
			instance.id,
			discoveryService,
			[]string{instance.endpoint},
			map[string]string{
				selector.MetadataRegion: "cn-east-1",
				selector.MetadataZone:   instance.zone,
			},
		)
		if err != nil {
			panic(fmt.Errorf("build instance %q: %w", instance.id, err))
		}
		if err := registryBackend.Register(ctx, value); err != nil {
			panic(fmt.Errorf("register instance %q: %w", instance.id, err))
		}
	}

	selection, err := selector.NewRoundRobin(
		"http",
		selector.WithLocality(selector.Locality{
			Region: "cn-east-1",
			Zone:   "zone-a",
		}),
	)
	if err != nil {
		panic(fmt.Errorf("build selector: %w", err))
	}
	router, err := client.NewRouter(client.RouterConfig{
		Name:      "example-discovery",
		Service:   discoveryService,
		Discovery: registryBackend,
		Selector:  selection,
		MaxStale:  10 * time.Second,
	})
	if err != nil {
		panic(fmt.Errorf("build discovery router: %w", err))
	}

	app, err := keelith.New(
		keelith.WithName("discovery-client"),
		keelith.WithHTTP(":8089"),
		keelith.WithComponent(router),
		keelith.WithRoute(http.MethodGet, "/pick", pick(router)),
	)
	if err != nil {
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}

func pick(router *client.Router) keelith.RouteHandler {
	return func(ctx context.Context, _ *http.Request) (any, error) {
		target, ok := operation.FromContext(ctx)
		if !ok {
			return nil, fmt.Errorf("request operation is missing")
		}
		node, done, err := router.Pick(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("pick discovery node: %w", err)
		}
		done(selector.Result{Latency: time.Millisecond})
		return map[string]any{
			"node":     node.ID(),
			"endpoint": node.Endpoint(),
			"state":    router.Describe().State,
		}, nil
	}
}
