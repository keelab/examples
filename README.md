# Keelith Examples

本仓库收集 Keelith 的渐进式 Go 示例。每个编号目录都是一个独立的 `main` 包，围绕一个运行时能力展开：从最小 HTTP 服务开始，逐步加入配置、生命周期、协议绑定、治理、后台任务、流式传输和依赖注入。

## 快速开始

### 环境要求

- Go 1.27 或更高版本
- Keelith 源码仓库（当前 `go.mod` 通过 `replace` 使用 `../keelith`）

推荐目录结构：

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

如果 Keelith 不在相邻目录，请修改 `go.mod` 末尾的 `replace` 路径。

### 基本约定

所有服务示例都从各自目录的 `main` 直接启动：创建信号上下文、构建应用、调用 `app.Run(ctx)`。启动阶段错误会立即终止示例；收到 `SIGINT` 或 `SIGTERM` 后的 `context.Canceled` 属于正常退出。示例使用本地端口，运行多个 HTTP 示例时请逐个启动。

## 示例索引

| 示例 | 主题 | 默认地址或运行方式 |
| --- | --- | --- |
| [01-hello-http](./01-hello-http) | 最小 HTTP 路由 | `:8080` · `GET /ping` |
| [02-business-routes](./02-business-routes) | 多路由与查询参数 | `:8080` · `/healthz`、`/greeting` |
| [03-file-config](./03-file-config) | 文件配置、严格字段与热加载 | `:8082` · `GET /message` |
| [04-component-lifecycle](./04-component-lifecycle) | Component 启停生命周期 | `:8083` · `GET /status` |
| [05-ops](./05-ops) | 独立 Ops 端口 | `:8084`、`:9090` |
| [06-service-profile](./06-service-profile) | Proto、Binding、Profile 与服务中间件 | `:8085` · `GET /v1/greeting` |
| [07-http-grpc-service](./07-http-grpc-service) | 同一服务的 HTTP + gRPC | `:8086`、`:8087` |
| [08-security-middleware](./08-security-middleware) | Metadata、认证、授权与错误映射 | `:8088` · `GET /whoami` |
| [09-discovery-client](./09-discovery-client) | 注册发现、选择器与客户端 Router | `:8089` · `GET /pick` |
| [10-worker-job](./10-worker-job) | Job、ACK 与优雅 drain | `:8090` · `POST /run` |
| [11-cache](./11-cache) | 类型化 read-through cache 与版本失效 | `:8091` · `/value`、`/invalidate` |
| [12-sse-stream](./12-sse-stream) | Server-Sent Events 与断点续传 | `127.0.0.1:18092` · `GET /events` |
| [13-websocket](./13-websocket) | WebSocket Hub 与双向流 | `127.0.0.1:8093` · `GET /ws` |
| [14-http-client](./14-http-client) | 类型化 HTTP 客户端与 outbound middleware | 本地 `httptest` |
| [15-continuation](./15-continuation) | Durable call 与可恢复状态迁移 | 内存 Store，一次性运行 |
| [16-topology-rollout](./16-topology-rollout) | Topology epoch、流量切换与 lease | 内存 Manager，一次性运行 |
| [17-di-graph](./17-di-graph) | 类型化 Provider、DI 图与清理 | `:8094` · `GET /greeting` |

## 运行示例

从仓库根目录执行：

```bash
# HTTP 示例
go run ./01-hello-http
curl http://127.0.0.1:8080/ping

# 配置热加载
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

# WebSocket（服务端与客户端需使用两个终端）
go run ./13-websocket
go run ./13-websocket/client
```

06、07 是 Proto 驱动示例，包含独立说明：

- [06-service-profile/README.md](./06-service-profile/README.md)：协议、生成代码、业务实现与 Profile 组装。
- [07-http-grpc-service/README.md](./07-http-grpc-service/README.md)：Buf 依赖、HTTP/gRPC 双监听与 typed client。

## 代码生成

06 使用 `protoc` 直调，07 使用 Buf。需要生成或修改 Proto 时安装：

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/keelab/keelith/cmd/protoc-gen-go-keelith@latest
```

生成目录中的 `*.pb.go`、`*.keelith.gen.go`、manifest 和 OpenAPI 文件均为生成产物，不要手工修改；应修改 `api/**/*.proto` 后重新生成。

## 验证与贡献

格式化并验证受影响的示例：

```bash
gofmt -w <changed-go-files>
go test ./<changed-package>
git diff --check
```

完整贡献约定见 [CONTRIBUTING.md](./CONTRIBUTING.md)。提交信息采用 Conventional Commits，例如 `docs: update examples guide`。

## 许可证

本项目以 [Apache License 2.0](./LICENSE) 发布。
