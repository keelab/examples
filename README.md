# Keelith Examples

English | [中文](./README.zh.md)

This repository contains progressive Go examples for Keelith. Each numbered directory is an independent `main` package focused on one runtime capability, starting with a minimal HTTP service and gradually adding configuration, lifecycle management, protocol bindings, governance, background jobs, streaming, and dependency injection.

## Quick start

### Requirements

- Go 1.27 or later
- A local checkout of the Keelith source repository (the current `go.mod` uses `replace ../keelith`)

Recommended layout:

```text
project/
├── keelith/
└── examples/
```

```bash
git clone https://github.com/keelab/keelith.git
git clone https://github.com/keelab/examples.git
cd examples
go mod download
go run ./01-hello-http
```

If Keelith is not in a neighboring directory, update the `replace` path at the end of `go.mod`.

### Conventions

Each service example starts from its directory's `main` package, creates a signal-aware context, builds the application, and calls `app.Run(ctx)`. Startup errors terminate the example immediately. `context.Canceled` after `SIGINT` or `SIGTERM` is a normal shutdown. Examples use local ports, so run HTTP examples one at a time.

## Example index

| Example | Topic | Default address or run mode |
| --- | --- | --- |
| [01-hello-http](./01-hello-http) | Minimal HTTP routing | `:8080` · `GET /ping` |
| [02-business-routes](./02-business-routes) | Multiple routes and query parameters | `:8080` · `/healthz`, `/greeting` |
| [03-file-config](./03-file-config) | File configuration and hot reload | `:8082` · `GET /message` |
| [04-component-lifecycle](./04-component-lifecycle) | Component lifecycle | `:8083` · `GET /status` |
| [05-ops](./05-ops) | Dedicated Ops port | `:8084`, `:9090` |
| [06-service-profile](./06-service-profile) | Proto, Binding, Profile, and middleware | `:8085` · `GET /v1/greeting` |
| [07-http-grpc-service](./07-http-grpc-service) | HTTP and gRPC for one service | `:8086`, `:8087` |
| [08-security-middleware](./08-security-middleware) | Metadata, authentication, authorization, and error mapping | `:8088` · `GET /whoami` |
| [09-discovery-client](./09-discovery-client) | Registration, discovery, selectors, and client Router | `:8089` · `GET /pick` |
| [10-worker-job](./10-worker-job) | Job, ACK, and graceful drain | `:8090` · `POST /run` |
| [11-cache](./11-cache) | Typed read-through cache and version invalidation | `:8091` |
| [12-sse-stream](./12-sse-stream) | Server-Sent Events and resumable streams | `127.0.0.1:18092` |
| [13-websocket](./13-websocket) | WebSocket Hub and bidirectional streams | `127.0.0.1:8093` |
| [14-http-client](./14-http-client) | Typed HTTP client and outbound middleware | Local `httptest` |
| [15-continuation](./15-continuation) | Durable calls and resumable state transitions | In-memory Store |
| [16-topology-rollout](./16-topology-rollout) | Topology epochs, traffic switching, and leases | In-memory Manager |
| [17-di-graph](./17-di-graph) | Typed providers, DI graph, and cleanup | `:8094` · `GET /greeting` |

## Running examples

Run from the repository root:

```bash
# HTTP
go run ./01-hello-http
curl http://127.0.0.1:8080/ping

# Configuration hot reload
go run ./03-file-config
curl http://127.0.0.1:8082/message

# Discovery
go run ./09-discovery-client
curl http://127.0.0.1:8089/pick

# Job
go run ./10-worker-job
curl -X POST http://127.0.0.1:8090/run

# SSE
go run ./12-sse-stream
curl -N http://127.0.0.1:18092/events

# WebSocket (server and client require two terminals)
go run ./13-websocket
go run ./13-websocket/client
```

Examples 06 and 07 are Proto-driven and include dedicated documentation:

- [06-service-profile/README.md](./06-service-profile/README.md): protocol, generated code, business implementation, and Profile assembly.
- [07-http-grpc-service/README.md](./07-http-grpc-service/README.md): Buf dependencies, dual HTTP/gRPC listeners, and the typed client.

## Code generation

Example 06 invokes `protoc` directly; example 07 uses Buf. Install the generators when creating or modifying Proto files:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/keelab/keelith/cmd/protoc-gen-go-keelith@latest
```

Generated `*.pb.go`, `*.keelith.gen.go`, manifest, and OpenAPI files must not be edited by hand. Modify `api/**/*.proto` and regenerate instead.

## Verification and contribution

```bash
gofmt -w <changed-go-files>
go test ./<changed-package>
git diff --check
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the complete guidelines. Use Conventional Commits, for example `docs: update examples guide`.

## License

This project is released under the [Apache License 2.0](./LICENSE).
