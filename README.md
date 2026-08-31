# Keelith 渐进示例

这些示例按概念递增排列。每一步都能独立运行，后一步只在确实需要时引入
前一步没有的运行时能力。

这些示例参考大型 Go 项目常见的采用路径：先用一个进程验证业务，再逐步加入配置、依赖生命周期和运维入口。以下命令均从仓库根目录执行，各服务目录可以独立启动，示例之间没有隐式启动顺序；07 的 `client` 目录需要先启动同级服务端。

## 这套目录为什么这样排

我实际阅读了各框架仓库的入口源码，而不是只按官网功能表抄目录：

- go-zero 的 `tools/goctl/example/rpc/hello/hello.go`、`hi/hi.go` 和
  `zero-examples/bookstore/api/bookstore.go` 把协议生成、`ServiceContext`、RPC 注册和
  API 组合根分开；RPC 生成器还单独覆盖多服务、导入链和 streaming。
- Kratos 的 `examples/helloworld/server/main.go` 是最小 HTTP + gRPC 双服务入口，
  `examples/blog` 再加入 data/biz/service 分层，`examples/registry`、`metrics`、
  `traces`、`validate` 分别展示注册发现、观测和校验。
- CloudWeGo/Hertz 的 `hello/main.go` 只创建一个 server，`middleware/*`、`multiple_service`、
  `graceful_shutdown`、`opentelemetry/*` 按一个能力一个目录递进。
- CloudWeGo/Kitex 的 `basic/server`、`discovery`、`governance/circuitbreak`、
  `streaming` 和 `opentelemetry` 把 IDL 生成、客户端、发现、治理、流式调用和观测拆成
  可单独运行的主题目录。
- Go Micro 的 `examples/hello-world/main.go` 从 `NewService → Init → Handle → Run`
  开始，`multi-service`、`auth`、`grpc-interop` 再分别扩展组合、认证和协议互操作。
- TarsGo 的 `examples/EchoClientServer`、`ContextTestServer`、`OpentelemetryServer`
  则按 `*.tars → 生成代码 → Servant → Client → Filter/Tracing` 展开。

所以 Keelith 也采用“最小运行时、业务边界、配置与生命周期、传输与治理、客户端与
后台运行时、流式协议、CLI 生成器”的顺序；需要 PostgreSQL、Redis、Kafka、注册中心
或 Kubernetes 的能力保留在后面的生产参考，不把外部环境伪装成最小示例的前置条件。
CloudWeGo 的 Hertz/Kitex 适配入口见 [`x/README.md`](../x/README.md)，外部基础设施适配见
[`contrib/README.md`](../contrib/README.md)。

## 示例代码约定

所有示例都从各自目录的 `main` 直接启动：先创建可取消的上下文，再构建应用组件，最后
调用 `app.Run(ctx)`。启动阶段的配置、依赖和路由错误会立即终止示例；收到
`SIGINT`/`SIGTERM` 后的 `context.Canceled` 属于正常退出。这样 01–17 的入口结构保持一致，
复杂示例仍将具体能力拆分到独立 helper 中。

## 01：最小 HTTP

```bash
go run ./01-hello-http
curl http://127.0.0.1:8080/ping
```

只需要 `keelith.New`、监听地址、一个 Route 和 `Run`。它不依赖配置文件、DI 图或外部服务。

## 02：多路由业务服务

```bash
go run ./02-business-routes
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/greeting?name=Keelith
```

这个例子展示一个小型业务服务如何把路由处理函数拆成独立函数，并处理查询参数和错误。它仍然只使用内存状态，适合先验证 HTTP 业务边界。

## 03：文件配置与热加载

```bash
go run ./03-file-config
curl http://127.0.0.1:8082/message
```

`configs/dev.yaml` 通过 `WithConfigFile` 接入。修改文件后，配置运行时会按轮询间隔重新加载；这个层级开始需要理解配置文件路径、严格字段和环境变量前缀，但仍不需要数据库或注册中心。

## 04：组件生命周期

```bash
go run ./04-component-lifecycle
curl http://127.0.0.1:8083/status
```

这里用 `app.ComponentFunc` 表达一个可启动、可停止的依赖。应用负责启动顺序、优雅停止和错误传播，业务代码不需要自己管理 goroutine 或信号。

## 05：Ops 运维入口

```bash
go run ./05-ops
curl http://127.0.0.1:8084/ping
curl http://127.0.0.1:9090/health/ready
curl http://127.0.0.1:9090/debug/build
```

Ops listener 默认只暴露健康检查；显式启用 `WithBuildInfo` 后才提供构建信息。生产服务通常从这一层开始接入独立运维端口，而不是把诊断接口混入业务路由。

## 06：Service Binding、Profile 与中间件

```bash
go run ./06-service-profile
curl http://127.0.0.1:8085/v1/greeting
```

这一层演示生成器产出的 `BindXxx` 组合根：业务实现满足
`GreetingServiceKeelithServer`，再由 `protoc-gen-go-keelith` 生成的
`BindGreetingService` 描述服务身份、HTTP 注册和服务级中间件，最后通过
`service.Profile` 组成一个可审计的部署单元。

**详细步骤**（如何写 `api/`、生成 `gen/`、实现 `internal/`）见
[`06-service-profile/README.md`](./06-service-profile/README.md)。

## 07：HTTP + gRPC 服务绑定

终端一启动服务：

```bash
go run ./07-http-grpc-service
```

终端二调用生成的 gRPC 客户端：

```bash
go run ./07-http-grpc-service/client
```

也可以先验证 HTTP：

```bash
curl http://127.0.0.1:8086/v1/ping
```

与 06 相同，Binding 由 `protoc-gen-go-keelith` 从 Proto 生成；本示例额外演示
HTTP 与 gRPC 双监听，以及 `NewPingServiceGRPCClient` 类型化客户端。

**详细步骤**见 [`07-http-grpc-service/README.md`](./07-http-grpc-service/README.md)。

## 08：认证、授权与错误映射

```bash
go run ./08-security-middleware
curl -i http://127.0.0.1:8088/whoami
curl -i -H 'Authorization: Bearer guest-token' http://127.0.0.1:8088/whoami
curl -i -H 'Authorization: Bearer reader-token' http://127.0.0.1:8088/whoami
```

依次可以看到 401、403 和 200。`authorization` 先经过 metadata allowlist，再由
AuthN 生成 Principal，最后由 RBAC 按 `operation.Operation` 判定；凭据不会写入日志
或响应。

## 09：注册发现与客户端选择

```bash
go run ./09-discovery-client
curl http://127.0.0.1:8089/pick
```

示例用内存 Registry 发布三个节点，`client.Router` 作为 App-owned component 订阅
完整快照，RoundRobin selector 按本地 zone preference 选择节点并上报完成结果。替换
Registry 实现即可接入 etcd、Consul 或 Kubernetes，不需要改业务 handler。

## 10：后台 Job 与优雅 drain

```bash
go run ./10-worker-job
curl -X POST http://127.0.0.1:8090/run
```

`worker.Job` 只依赖 Scheduler 合约；示例提供一个手动触发的内存 scheduler，展示
`Schedule → ACK → StopPulling → Drain → Close` 生命周期。真实 Cron、Kafka 或平台
调度器可以替换这个 adapter。

## 11：类型化缓存与版本失效

```bash
go run ./11-cache
curl http://127.0.0.1:8091/value
curl http://127.0.0.1:8091/value
curl -X POST http://127.0.0.1:8091/invalidate
```

两次读取只触发一次 loader，失效接口使用 versioned backend 的单调版本；内存 backend
只用于说明 Cache contract，生产环境应替换为 Redis 等受管适配器。

## 12：Server-Sent Events

```bash
go run ./12-sse-stream
curl -N http://127.0.0.1:18092/events
curl -N -H 'Last-Event-ID: 1' http://127.0.0.1:18092/events
```

这个例子直接使用与生成代码相同的 `DecodeSSERequest`、`NewSSEEncoder` 和
`WithStreaming`，并演示 Last-Event-ID 断点续传。

## 13：WebSocket 双向流

终端一启动服务：

```bash
go run ./13-websocket
```

终端二运行内置客户端：

```bash
go run ./13-websocket/client
```

也可以使用 `websocat` 手动连接：

```bash
websocat ws://127.0.0.1:8093/ws
```

`websocket.Hub` 由 App 管理连接上限、握手、消息预算和停止时的连接 drain；
`StreamBundle` 只记录 stream lifecycle，不接触业务 payload。

## 14：HTTP 客户端调用

```bash
go run ./14-http-client
```

使用 `transport/http.Client`、`ClientCall[T]` 和 outbound middleware 调用一个本地
`httptest` 服务，展示 metadata 注入、稳定 operation identity 和 typed response decode。

## 15：可恢复的 Continuation

```bash
go run ./15-continuation
```

这个例子把一次调用写入内存 Store，再由冻结后的 Machine Registry 执行一次带 fencing
的状态迁移。它展示 durable call、revision 和完成 frame；生产环境只需替换持久化 Store
与调度入口。

## 16：Topology epoch 与滚动切换

```bash
go run ./16-topology-rollout
```

先激活 blue epoch，再以 green epoch 接管全部流量，最后 drain/stop 旧 epoch。示例只用
内存 Manager，重点是不可变 plan、依赖绑定解析、traffic weight 和 call lease 的顺序。

## 17：类型化 DI 图

```bash
go run ./17-di-graph
curl http://127.0.0.1:8094/greeting
```

`di.Provide` 和 `di.Build` 负责构造带类型的依赖图，`WithGraph` 将图的关闭交给
Application。业务处理函数只接收已经构造好的对象，不需要持有 Container 或依赖反射。

## CLI 生成的标准服务

```bash
keelith new orders --template service --module example.com/orders
cd orders
go run .
```

标准模板把协议、Binding、业务 Handler 和 Keelith CLI 生成的 Application 组合根分开；需要检查对象图时执行：

```bash
keelith wiring sync
keelith wiring check
keelith wiring verify
keelith wiring graph
keelith graph
```

## 生产参考

```bash
keelith new platform --template production --module example.com/platform
```

这个模板用于学习显式 Profile、Group、严格配置和 wiring 合同，不会自动连接
PostgreSQL、Redis、Kafka 或注册中心。需要真实 Outbox、Topology、外部组件和部署
验收时，请继续阅读独立 Demo 仓库中的完整生产参考。

## 建议的扩展顺序

```text
最小 HTTP → 业务路由 → 配置 → Component → Ops
→ Binding/安全 → Proto + gRPC → Discovery/Job/Cache
→ SSE/WebSocket/Client → Continuation/Topology/DI → 数据库/Outbox
```
