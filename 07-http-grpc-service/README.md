# 07：HTTP + gRPC 服务绑定示例

在 [06-service-profile](../06-service-profile/README.md) 的基础上，本示例展示**同一个 Proto 服务同时挂载 HTTP 和 gRPC 监听**，并使用生成的 gRPC 客户端调用。

## 运行

在仓库根目录（`examples/`）执行。

终端一，启动服务：

```bash
go run ./07-http-grpc-service
```

终端二，验证 HTTP：

```bash
curl http://127.0.0.1:8086/v1/ping
# {}
```

终端三，验证 gRPC（使用生成的 typed client）：

```bash
go run ./07-http-grpc-service/client
# gRPC response: pong
```

`service-log` 中间件日志会打印在**服务端终端**（HTTP 和 gRPC 请求各打一行）。

## 目录与文件来源

```text
07-http-grpc-service/
├── api/ping/v1/ping.proto           # 手写：协议源文件
├── buf.yaml / buf.gen.yaml / buf.lock
├── gen/ping/v1/
│   ├── ping.pb.go                   # 生成：protoc-gen-go
│   ├── ping.keelith.gen.go          # 生成：protoc-gen-go-keelith
│   ├── ping.keelith.manifest.json
│   └── ping.openapi.json
├── internal/ping/service.go         # 手写：业务实现
├── client/main.go                   # 手写：gRPC 客户端演示
└── main.go                          # 手写：组装 Profile 并启动双监听
```

| 路径 | 谁维护 | 作用 |
|------|--------|------|
| `api/**/*.proto` | 开发者手写 | 定义 RPC 与 `google.api.http` 路由 |
| `gen/**/*.keelith.gen.go` | `protoc-gen-go-keelith` 生成 | `PingServiceKeelithServer`、`BindPingService`、`NewPingServiceGRPCClient` |
| `internal/ping/` | 开发者手写 | 实现 `Ping` 业务逻辑 |
| `main.go` | 开发者手写 | `WithHTTP` + `WithGRPC` + `WithProfile` |
| `client/main.go` | 开发者手写 | 演示如何用生成客户端调用 gRPC |

与 06 的区别：本示例在 `BindPingService` 上同时挂了 `WithHTTPBundle` 和 `WithGRPCBundle`，并启用 `keelith.WithGRPC(":8087")`。

## 前置工具

本示例用 [Buf](https://buf.build/docs/cli/installation/) 管理 Proto 依赖与代码生成（06 示例仍用 `protoc` 直调，可按需对照）。

```bash
brew install bufbuild/buf/buf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/keelab/keelith/cmd/protoc-gen-go-keelith@latest
```

`ping.proto` 中的 `google/api/annotations.proto` 来自 `buf.yaml` 声明的依赖 `buf.build/googleapis/googleapis`。首次 clone 或修改 `deps` 后，需要先拉取依赖：

```bash
cd 07-http-grpc-service
buf dep update   # 生成 buf.lock，下载 googleapis
```

若跳过此步直接 `buf generate`，会报错：

```text
imported file does not exist
cannot find `google.api.http` in this scope
```

IDE 中 `import` 报红时，安装 [Buf VS Code 扩展](https://buf.build/docs/editor-integration) 并执行 `buf dep update`。

## 步骤 1：编写 API（`api/ping/v1/ping.proto`）

```protobuf
syntax = "proto3";

package ping.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1;examplesv1";

service PingService {
  rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      get: "/v1/ping"
    };
  }
}
```

- `google.api.http` 生成 HTTP 路由 `GET /v1/ping`
- 无 HTTP 注解的 RPC 只生成 gRPC 适配
- `package` 使用 `ping.v1`（不要用 `examples.v1`），避免与 `keelith` 仓库内同名示例的 proto 注册冲突
- `option go_package` 的目录段应与 Proto 路径一致（`api/ping/v1` → `gen/ping/v1`）

## 步骤 2：生成 `gen/` 代码

在 `07-http-grpc-service` 目录下执行：

```bash
cd 07-http-grpc-service
buf dep update   # 首次或 deps 变更后执行
buf generate
```

**不要手改 `gen/` 下的生成文件**；改 Proto 后重新执行 `buf generate` 即可。

生成文件 `ping.keelith.gen.go` 提供：

- `PingServiceKeelithServer`：服务端接口
- `BindPingService(...)`：HTTP + gRPC 双传输 Binding
- `NewPingServiceGRPCClient(...)`：类型化 gRPC 客户端

## 步骤 3：编写业务实现（`internal/ping/service.go`）

```go
package ping

import (
	"context"

	examplesv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (*Service) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

var _ examplesv1.PingServiceKeelithServer = (*Service)(nil)
```

## 步骤 4：在 `main.go` 组装双监听 Profile

```go
import (
	examplesv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
	"github.com/keelab/examples/07-http-grpc-service/internal/ping"
)

profile, err := service.NewProfile(
	"ping",
	examplesv1.BindPingService(
		ping.New(),
		service.WithHTTPBundle(bundle),
		service.WithGRPCBundle(bundle),
	),
)

application, err := keelith.New(
	keelith.WithName("http-grpc-service"),
	keelith.WithHTTP(":8086"),
	keelith.WithGRPC(":8087"),
	keelith.WithProfile(profile),
)
```

## 步骤 5：客户端使用生成代码（`client/main.go`）

```go
import (
	examplesv1 "github.com/keelab/examples/07-http-grpc-service/gen/ping/v1"
)

client := examplesv1.NewPingServiceGRPCClient(connection)
_, err := client.Ping(ctx, &emptypb.Empty{})
```

客户端走生成的 typed API，而不是手写 `ServiceDesc` 或依赖 gRPC reflection。

## 常见问题：`proto name conflict`

若启动时报：

```text
panic: proto: file "examples/v1/ping.proto" has a name conflict over examples.v1.PingService
```

说明同一进程里**同时链接了两套 gen 代码**（本仓库的 `gen/ping/v1` 与 `keelith` 仓库内的 `examples/07-http-grpc-service/gen/examples/v1`）。

处理方式：

1. 确认 `main.go`、`internal/`、`client/` 的 import 都指向 `github.com/keelab/examples/07-http-grpc-service/gen/ping/v1`，不要引用 `github.com/keelab/keelith/examples/...`
2. 本示例 Proto 的 `package` 应为 `ping.v1`（已规避与 keelith 内置示例的命名冲突）
3. 修改 Proto 后执行 `buf generate` 重新生成 `gen/`
4. GoLand 中 **File → Invalidate Caches** 后重新构建

## 数据流

```text
api/ping/v1/ping.proto
    → buf dep update + buf generate
        → gen/ping/v1/*.pb.go + *.keelith.gen.go
            → BindPingService + PingServiceKeelithServer + GRPCClient
                → internal/ping.Service 实现服务端
                → main.go WithProfile 挂载 HTTP:8086 + gRPC:8087
                → client/main.go 用 NewPingServiceGRPCClient 调用
```
