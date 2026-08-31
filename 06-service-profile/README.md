# 06：Service Profile 示例

这个示例演示 Keelith 推荐的服务组织方式：**协议在 `api/`，生成代码在 `gen/`，业务实现在 `internal/`，启动入口在 `main.go`**。Binding 不由手写，而是由 `protoc-gen-go-keelith` 从 Proto 生成。

## 运行

在项目根目录执行：

```bash
go run ./examples/06-service-profile
```

另开终端验证：

```bash
curl http://127.0.0.1:8085/v1/greeting
# {"message":"hello from a profile binding"}
```

`request-log` 中间件的日志会打印在**运行服务的终端**，不会出现在 `curl` 输出里。

## 目录与文件来源

```text
examples/06-service-profile/
├── api/greeting/v1/greeting.proto   # 手写：协议源文件
├── gen/greeting/v1/
│   ├── greeting.pb.go               # 生成：protoc-gen-go
│   ├── greeting.keelith.gen.go      # 生成：protoc-gen-go-keelith
│   ├── greeting.keelith.manifest.json
│   └── greeting.openapi.json
├── internal/greeting/service.go     # 手写：业务实现
└── main.go                          # 手写：组装 Profile 并启动
```

| 路径 | 谁维护 | 作用 |
|------|--------|------|
| `api/**/*.proto` | 开发者手写 | 定义消息、RPC、HTTP 路由注解 |
| `gen/**/*.pb.go` | `protoc-gen-go` 生成 | Protobuf 消息类型 |
| `gen/**/*.keelith.gen.go` | `protoc-gen-go-keelith` 生成 | `GreetingServiceKeelithServer` 接口、`BindGreetingService`、HTTP/gRPC 注册 |
| `gen/**/*.keelith.manifest.json` | `protoc-gen-go-keelith` 生成 | 服务契约清单，供 CLI / 依赖图使用 |
| `internal/<svc>/` | 开发者手写 | 实现生成接口中的业务逻辑 |
| `main.go` | 开发者手写 | 创建中间件、Profile、`keelith.New` |

正式项目里也可以用 `keelith new` / `keelith add service` 脚手架生成同样布局；本示例把每一步展开，方便对照理解。

## 前置工具

```bash
# protoc（macOS 可用 brew install protobuf）
protoc --version

# Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/keelab/keelith/cmd/protoc-gen-go-keelith@latest

# google.api.http 注解依赖的 Proto 描述文件（一次性导出）
buf export buf.build/googleapis/googleapis --output /tmp/googleapis
```

确保 `$(go env GOPATH)/bin` 在 `PATH` 中，以便 `protoc` 能找到 `protoc-gen-go` 和 `protoc-gen-go-keelith`。

## 步骤 1：编写 API（`api/`）

在 `api/greeting/v1/greeting.proto` 中声明服务。关键点：

- `package` 决定逻辑服务名（如 `greeting.v1.GreetingService`）
- `option go_package` 决定生成代码的 Go import 路径和包名
- `google.api.http` 注解声明 HTTP 路由；没有它则只生成 gRPC 适配

```protobuf
syntax = "proto3";

package greeting.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/keelab/keelith/examples/06-service-profile/gen/greeting/v1;greetingv1";

message GetGreetingResponse {
  string message = 1;
}

service GreetingService {
  rpc GetGreeting(google.protobuf.Empty) returns (GetGreetingResponse) {
    option (google.api.http) = {
      get: "/v1/greeting"
    };
  }
}
```

新增 RPC 时：在 `service` 块里加方法、补充 request/response message，并写上对应的 `google.api.http` 注解。

## 步骤 2：生成 `gen/` 代码

在 `examples/06-service-profile` 目录下执行：

```bash
cd examples/06-service-profile

PROTOBUF_INCLUDE="$(go list -m -f '{{.Dir}}' google.golang.org/protobuf)/types/known/.."
GOOGLEAPIS=/tmp/googleapis   # 见上文 buf export

mkdir -p gen/greeting/v1

protoc \
  -I api/greeting/v1 \
  -I "$PROTOBUF_INCLUDE" \
  -I "$GOOGLEAPIS" \
  --go_out=gen/greeting/v1 \
  --go_opt=paths=source_relative \
  --go-keelith_out=gen/greeting/v1 \
  --go-keelith_opt=paths=source_relative \
  api/greeting/v1/greeting.proto
```

生成后重点关注 `greeting.keelith.gen.go`，它会提供：

- `GreetingServiceKeelithServer`：业务侧需要实现的接口
- `BindGreetingService(...)`：把实现包装成 `service.Binding`
- `RegisterGreetingServiceHTTP` / `RegisterGreetingServiceGRPC`：传输层注册函数

**不要手改 `gen/` 下的生成文件**；改 Proto 后重新执行上面的 `protoc` 命令即可。

## 步骤 3：编写业务实现（`internal/greeting/service.go`）

业务代码只依赖生成接口，不直接操作 HTTP Router 或 gRPC `ServiceDesc`：

```go
package greeting

import (
	"context"

	greetingv1 "github.com/keelab/examples/06-service-profile/gen/greeting/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct{}

func New() *Service { return &Service{} }

func (*Service) GetGreeting(
	_ context.Context,
	_ *emptypb.Empty,
) (*greetingv1.GetGreetingResponse, error) {
	return &greetingv1.GetGreetingResponse{Message: "hello from a profile binding"}, nil
}

// 编译期断言：确保实现了生成接口
var _ greetingv1.GreetingServiceKeelithServer = (*Service)(nil)
```

约定：

- 方法签名与 Proto RPC 一一对应（参数、返回值类型来自 `gen/`）
- 通过 `var _ XxxKeelithServer = (*Service)(nil)` 保证接口变更时编译失败
- 一个 `internal/<name>/` 目录对应一个业务服务；不要把传输层代码写在这里

## 步骤 4：在 `main.go` 组装 Profile

`main.go` 负责把生成 Binding、业务实现、中间件和 Keelith 应用串起来：

```go
profile, err := service.NewProfile(
	"greeting",
	greetingv1.BindGreetingService(
		greeting.New(),
		service.WithHTTPBundle(bundle), // 服务级 HTTP 中间件
	),
)

application, err := keelith.New(
	keelith.WithName("service-profile"),
	keelith.WithHTTP(":8085"),
	keelith.WithProfile(profile),
)
```

数据流：

```text
greeting.proto
    → protoc-gen-go-keelith
        → BindGreetingService + KeelithServer 接口
            → internal/greeting.Service 实现接口
                → service.NewProfile 收集 Binding
                    → keelith.WithProfile 挂载到 HTTP 监听
```

## 新增一个 RPC 的完整流程

1. 编辑 `api/greeting/v1/greeting.proto`，增加 message 和 rpc（含 `google.api.http`）
2. 重新执行步骤 2 的 `protoc` 命令
3. 在 `internal/greeting/service.go` 实现生成接口新增的方法
4. 若需中间件，继续在 `BindGreetingService` 的 `service.WithHTTPBundle` / `WithGRPCBundle` 上挂载
5. `go run` 验证

## 与正式项目的对应关系

| 本示例 | `keelith new --template service` 脚手架 |
|--------|----------------------------------------|
| `api/greeting/v1/greeting.proto` | `api/echo/v1/echo.proto` |
| `gen/greeting/v1/*.keelith.gen.go` | `gen/echo/v1/*.keelith.gen.go` |
| `internal/greeting/service.go` | `internal/echo/service.go` |
| `main.go` 手写 `WithProfile` | `internal/keelithgen/application.keelith.gen.go` 由 CLI wiring 生成 |

脚手架会把 `main.go` 里的组装逻辑下沉到 `internal/keelithgen/`，但分层原则相同：**Proto 驱动生成，internal 只写业务，Binding 不手写**。
