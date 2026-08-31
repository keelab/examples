// Command security-middleware demonstrates metadata policy, authentication,
// authorization and transport-neutral application errors.
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
	"github.com/keelab/keelith/metadata"
	"github.com/keelab/keelith/middleware"
	"github.com/keelab/keelith/security"
	"github.com/keelab/keelith/security/authn"
	"github.com/keelab/keelith/security/authz"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reader, _ := security.NewPrincipal(security.PrincipalSpec{
		Subject: "user-reader",
		Issuer:  "example",
		Roles:   []string{"reader"},
	})
	guest, _ := security.NewPrincipal(security.PrincipalSpec{
		Subject: "user-guest",
		Issuer:  "example",
		Roles:   []string{"guest"},
	})

	bearer, err := authn.NewBearer(map[string]security.Principal{
		"reader-token": reader,
		"guest-token":  guest,
	})
	if err != nil {
		panic(err)
	}

	extractor, _ := authn.MetadataBearer("authorization", 4096)
	authenticate, _ := authn.Middleware(extractor, bearer)
	rbac, err := authz.NewRBAC(authz.Rule{
		Service: "keelith.quickstart",
		Method:  "GET /whoami",
		AnyRole: []string{"reader"},
	})
	if err != nil {
		panic(err)
	}
	authorize, err := authz.Middleware(rbac)
	bundle, err := middleware.NewServerBundle(middleware.ServerBundleConfig{
		Source: "examples/08-security-middleware",
		AdditionalEntries: []middleware.Entry{
			{Name: "authentication", Middleware: authenticate},
			{Name: "authorization", Middleware: authorize},
		},
	})
	policy, err := metadata.NewPolicy(
		[]string{"authorization"},
		metadata.WithSensitiveKeys("authorization"),
	)
	app, err := keelith.New(
		keelith.WithName("security-middleware"),
		keelith.WithHTTP(":8088"),
		keelith.WithMetadataPolicy(policy),
		keelith.WithMiddleware(bundle),
		keelith.WithRoute(http.MethodGet, "/whoami", whoami),
	)
	if err != nil {
		panic(err)
	}

	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}

func whoami(ctx context.Context, _ *http.Request) (any, error) {
	principal, ok := security.PrincipalFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("principal was not attached")
	}
	return map[string]any{
		"subject": principal.Subject(),
		"roles":   principal.Roles(),
	}, nil
}
