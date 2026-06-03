---
alwaysApply: true
description: 项目规范 (Project Rules)
---
# 项目规范 (Project Rules)

本文档定义了项目的开发规范，旨在确保代码的一致性和可维护性。AI 生成代码时必须严格遵守此规范。

# 项目介绍
本项目是提供hertz服务路由的im chat服务

## 最重要的前提
- 思考代码的可维护性、可扩展性、可复用性。
- 避免重复代码，保持代码的 DRY (Don't Repeat Yourself) 原则。
- 如果能抽象出来一个工具或者组件的，就抽象出来一个工具或者组件。
- 严禁写出代码屎山，即代码重复、逻辑复杂、可维护性差的代码。
- 不用过度设计， 大道至简。
- 注重代码性能，可用零拷贝就用，可用池化就用
- 每个业务都要考虑安全退出的机制

## 1. 技术栈概览
- **语言**: Go 1.26+
- **依赖注入**: Google Wire
- **日志库**: Slog (Structured Logging)
- **redis** go-redis
- **postgres** sqlx
- **grpc** 调用github.com/project-kgo/kim-gate提供长连接相关能力
- **json** 使用sonic库进行json编码解码

## 2. 编码规范
- **日志**: 使用 `log/slog` 进行结构化日志记录，通过依赖注入传入 Logger。
- **依赖注入**: 使用 `google/wire` 进行依赖注入。修改依赖关系后需重新生成 `wire_gen.go`

## 3. 代码结构
项目采用单服务分层结构，入口、依赖装配、业务分层和协议文件职责必须清晰分离。

```text
.
├── main.go                 # 服务入口：加载配置、初始化应用、监听退出信号、执行安全关闭
├── wire.go                 # Wire 依赖声明，新增依赖关系时在此维护
├── wire_gen.go             # Wire 生成文件，禁止手改
├── internal/               # 项目内部代码，禁止被外部模块直接引用
│   ├── app/                # 应用编排层：组装 HTTP/RPC/消费者，统一启动与关闭
│   ├── config/             # 配置加载、解析、默认值和校验
│   ├── data/               # 数据访问层：Postgres、Redis、MQ Redis、去重存储等
│   ├── discovery/          # 服务发现抽象
│   │   └── etcd/           # etcd registry/resolver 实现
│   ├── event/              # 事件主题、消息事件结构等异步消息定义
│   ├── gateway/            # kim-gate gRPC 客户端封装
│   ├── handler/            # Hertz HTTP handler，只做入参解析、鉴权上下文和响应封装
│   ├── middleware/         # Hertz 中间件
│   ├── model/              # 请求、响应、领域模型和数据传输结构
│   ├── router/             # HTTP 路由注册，统一管理路径和 handler 绑定
│   ├── rpc/                # 本服务暴露的 RPC Server
│   └── service/            # 业务逻辑层：编排 data、gateway、event，不直接处理 HTTP 细节
├── proto/                  # Protobuf 协议定义及生成代码
│   └── kim/v1/             # kim 服务 v1 协议
└── docs/                   # 项目文档、SQL、设计文档和执行计划
    └── sql/                # 数据库表结构和迁移参考 SQL
```

### 分层约束
- `handler` 只负责协议层适配：解析请求、调用 `service`、返回统一响应；不要在 handler 中直接访问数据库、Redis 或 gRPC gateway。
- `service` 负责业务规则、幂等、事件发布、跨资源编排；需要持久化或缓存时通过 `data` 层提供的方法完成。
- `data` 是唯一允许直接操作 Postgres、Redis、MQ Redis 的层；SQL、Redis key、数据读写细节不要散落到其它层。
- `gateway` 只封装对 `github.com/project-kgo/kim-gate` 的 gRPC 调用和连接生命周期；业务层通过接口依赖它，便于测试。
- `router` 只做路由注册和中间件挂载；不要承载业务逻辑。
- `model` 放稳定的数据结构；不要在其中引入具体数据库、HTTP、gRPC 客户端依赖。
- `event` 放事件结构、topic、content-type 等异步消息协议；发布和消费逻辑放在 `service`。
- `app` 统一管理启动、停止和资源释放；新增后台消费者、RPC 服务、长连接资源时必须接入 `App.Shutdown`。

### 新增代码放置规则
- 新增 HTTP API：优先在 `model` 定义请求/响应，在 `handler` 增加入口，在 `service` 实现业务，在 `router` 注册路由。
- 新增数据库或 Redis 操作：在 `internal/data` 新增 store/repository 方法，并提供可测试的接口或方法；不要从 `service` 直接拼 SQL 或 Redis key。
- 新增外部服务调用：在独立包中封装 client、配置和关闭逻辑，再通过 Wire 注入到 `service`。
- 新增 protobuf：修改 `proto/kim/v1/*.proto` 后重新生成对应 `.pb.go`，生成文件不手工修改。
- 新增依赖：更新 `wire.go` 后执行 Wire 生成，保持 `wire_gen.go` 与依赖图一致。
- 新增测试：测试文件与被测文件同目录，命名为 `*_test.go`；业务逻辑、data store、路由和生命周期逻辑都应优先补测试。

## 4. 注意事项
- 文件命名保持一致，避免拼写错误。
- 所有关于数据库/redis的操作都在 data层进行
