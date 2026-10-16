# NumberMan1/log

[![Go CI](https://github.com/NumberMan1/log/actions/workflows/go.yml/badge.svg)](https://github.com/NumberMan1/log/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/NumberMan1/log.svg)](https://pkg.go.dev/github.com/NumberMan1/log)

通用 Go 结构化日志库，模块路径为 `github.com/NumberMan1/log`，根包名为 `serverlog`。
要求 Go 1.25 或以上版本，不依赖私有业务模块。

库提供等级路由、字段 scope、标准 context、控制台／文件／Sentry 输出。
业务字段、用户身份、部署地域、请求数据和追踪信息由调用方显式传入。

## 安装

在使用方的 Go module 中执行：

```sh
go get github.com/NumberMan1/log@latest
```

可将 `@latest` 替换为 [已有 tag](https://github.com/NumberMan1/log/tags) 以固定版本。
API 见 [Go 文档](https://pkg.go.dev/github.com/NumberMan1/log)，升级前请阅读下方兼容说明与 [变更记录](CHANGELOG.md)。

## 快速开始

```go
package main

import (
    "context"
    "errors"

    serverlog "github.com/NumberMan1/log"
    "github.com/NumberMan1/log/field"
)

func main() {
    if err := serverlog.Configure(serverlog.Config{
        Level:       "info",
        Stdout:      true,
        StdoutTyp:   "json",
        LogFilePath: "./logs", // 留空则关闭文件输出
        Fields:      serverlog.M{"component": "example"}, // 调用方决定字段名
    }); err != nil {
        panic(err)
    }
    defer serverlog.Close()

    logger := serverlog.DefaultLogger().With(field.String("operation", "load"))
    ctx := serverlog.WithContext(context.Background(), logger)
    ctx = serverlog.WithFields(ctx, field.Int("attempt", 1))
    serverlog.MustFromContext(ctx).Error("operation failed",
        field.Error(errors.New("resource unavailable")),
    )
}
```

`L` 是管理输出与关闭的底层 core，使用 `Log(level, msg, M)`。
`Logger` 是带字段的接口，使用 `With`、`Clone` 和可变数量的 `field.Field`。
字段 scope 共享底层 core，不独立关闭输出资源。

## 配置与等级

| 配置 | 含义 |
|---|---|
| `Level` | 优先于 `MAX_LOG_LEVEL`；都为空时，release 构建为 info，其余为 debug |
| `Stdout` / `StdoutTyp` | 是否输出到终端；格式为 text（默认）或 json |
| `LogFilePath` | 直接写入的轮转根目录，空值禁用；不会创建隐式 log/event 子目录 |
| `Fields` | 显式公共字段，构造时复制顶层 map，本次调用可覆盖同名字段 |
| `SentryDSN` | 启用 Sentry，并覆盖高级配置中的 DSN |
| `SentryEnvironment` / `SentryRelease` | 非空时覆盖高级配置中的环境与版本 |
| `Sentry` | 其余 SDK、字段映射、Transport、故障回调配置 |

配置创建的 core 补充通用元数据 `msg`、`level`、`date`、`caller`，保留显式传入的 caller。
不会推导 app、server、region 或用户字段。文件输出位于轮转器决定的日期目录中。
Sentry 输出只启用 Error/Fatal/Panic，其余输出允许配置范围内的所有等级。

`Configure` 在所有输出初始化成功后才替换默认 core；失败时保留原默认值。
替换后会关闭旧 core，已有 scope 仍持有旧引用，应重新获取默认 logger。
独立实例使用 `NewConfiguredLogger(Config)`，由调用方负责 Close。
旧的 `Init(logPath, sentryDSN)` 仅成功初始化一次，初始化错误会 panic。

等级按严重程度递减：panic、fatal、error、warn/warning、info、debug、trace。
直接 `NewLogger()` 创建的 core 默认允许 Trace，不安装输出器或自动补充字段。
`Fatal` 在关闭当前及默认 core 后以状态码 1 退出，即使输出被过滤也会退出；`Panic` 输出后 panic。
非法配置返回错误；旧的包初始化逻辑仍会在非法 `MAX_LOG_LEVEL` 下 panic。

## 字段与 context

`field` 提供基础类型、切片、`Any`、`Object`、`Error` 及通用 trace/span 字段构造函数。
`field.Error` 保留原始 error，不解包业务错误。`Any`/`Object` 不提前序列化数据。

```go
parent := serverlog.DefaultLogger().With(field.String("component", "example"))
child := parent.Clone().With(field.String("operation", "load"))
child.Info("started", field.Int("attempt", 1))
```

`With`、`Clone` 复制字段切片，同名字段以后写入的值为准。
scope、日志调用及 Printer 分发使用独立顶层字段 map；嵌套 map、slice 和对象按只读 payload 使用。

- `WithContext(ctx, logger)` 返回标准派生 context，调用方须保留返回值。
- `WithFields(ctx, fields...)` 派生新的 context 和字段 scope，不修改父 context。
- `FromContext(ctx)` 返回 `(Logger, bool)`，nil 或缺失时返回 false。
- `MustFromContext(ctx)` 缺失时回退到当前默认 logger。
- `LoggerFromContext(ctx)` 缺失时返回 `ErrNoLoggerFound`。

标准 context 不可原地修改。框架 context 适配、追踪数据提取和业务错误规则在调用方实现。

## Sentry

```go
printer, err := serverlog.NewSentryPrinterWithError(serverlog.SentryPrinterConfig{
    Dsn:           "https://public@sentry.example.com/1",
    Environment:   "production",
    Release:       "v1.0.0",
    TagKeys:       []string{"component"}, // 显式选择需要索引的字段
    UserIDKeys:    []string{"identity"},  // 字段名称完全由调用方决定
    ExcludeKeys:   []string{"local_debug"},
})
if err != nil { panic(err) }
serverlog.AddPrinter(serverlog.LevelSwitchLte(serverlog.ErrorLevel), printer)
```

默认只把 `error`/`err` 中的原始 error 映射成 Exception，保留 unwrap 链和栈信息。
没有原始 error 时作为普通事件发送。默认不映射 tag、不推导用户、不排除字段，
也不按字段名删掉空值。传入的 `pid`、`user_id`、`request`、`response` 等仅是普通附加数据。

`TagKeys`、`UserIDKeys`、`ExcludeKeys` 均为显式选择；用户字段按顺序取首个非空、非零值。
`ExceptionKeys` 为 nil 时采用 error/err，非 nil 空切片禁用异常映射。
已映射字段不在附加数据中重复保存；字符串按 `MaxFieldLength` 裁剪，默认 4096 字节，保持有效 UTF-8。
附加 payload 保留结构化 map/slice，最多嵌套 16 层，每个集合最多 1000 项。

通过 `FieldMapper` 可完全替代默认映射。
`SentryFields.Extra` 在 sentry-go v0.47 中写入 `event.Contexts["extra"]`，保留每个字段结构；
该 SDK 版本已移除 `Event.Extra`。

每个 Printer 使用独立 Client/Scope，不继承 SDK 全局 hub 的 user/tags。
高级 SDK 选项通过 `ClientOptions` 传入；`Transport` 可注入离线测试实现。
环境与版本只采用显式配置和 Sentry 标准环境变量 `SENTRY_ENVIRONMENT`、`SENTRY_RELEASE`，
不读取应用自定义环境变量，也不内置 develop/production 环境名称。

`NewSentryPrinter` 保留配置错误时 panic 的构造行为；需要错误返回时使用 `NewSentryPrinterWithError`。
`Close` 限时 Flush 后释放 Transport。mapper/提交 panic 与 Flush 超时通过 `OnError`/`Err()` 报告，
SDK 后台网络重试状态由 Transport 管理。

## 文件和独立数据流

```go
// 需要第二条数据流时，调用方显式创建并管理独立 core。
dataLogger := serverlog.NewLogger()
printer, err := serverlog.NewDataPrinter("./data")
if err != nil { panic(err) }
dataLogger.AddPrinter(serverlog.LevelSwitchLte(serverlog.InfoLevel), printer)
defer dataLogger.Close()
dataLogger.Info("", serverlog.M{"value": 123})
```

`NewDataPrinter` 仅为数据补充 Time；`NewJsonlogPrinter` 包装消息、等级、时间及调用位置。
`NewFilePrinter` 可完全定制格式，默认只把 data 写成 JSON，不补充消息或等级。
应用日志配置默认使用 `StacksWrapPreHandler`；可通过底层 Printer/Builder 选择其他格式。

`FilePrinterConfig` 的默认缓冲为 4096 字节、队列为 2000 条、刷新间隔为 2 秒，按小时轮转。
可配置 `fileutil.DAY_SWITCHER` 或自定义 Switcher。

- `BlockOnFull` 默认等待队列；`DropNewest` 丢弃新日志并报告 `ErrLogQueueFull`，通过 `Dropped()` 查询累计数量。
- 写入、刷新、轮转和关闭失败通过 `OnError`/`Err()` 报告；默认写标准日志，不会后台 panic。
- 新文件独占创建，避免多个实例覆盖；新文件打开失败时保留旧文件。
- `Close` 等待已接受的数据排空；持续存储故障仍可能丢失数据，应监控错误并处理存储问题。
- 错误回调同步执行，需要快速返回，不能重入当前 Printer 或关闭其所属 core。

`JsonFormat` 在输出边界将嵌套 error 转成字符串，并在序列化失败时输出有效 JSON。
`fileutil.NewFile` 追加已有文件，`NewExclusiveFile` 预留新文件名，Close 刷新失败也会释放描述符。

## JSON 格式化与性能

`JsonFormat` 对基础类型、`map[string]any`、`[]any` 和常见基础类型切片使用类型分支，
仅在子项需要转换或截断时复制对应的通用 map/slice；无需转换的分支按只读数据复用。
命名容器与其他类型保留反射兜底，JSON 编码仍使用标准库，不改变原始 error 或输入字段。
调用方在格式化期间不得并发修改 payload；同一份只读 payload 可以并发记录。

map/slice/array 的递归转换在深度达到 32 时以 `"[max depth]"` 代替非 nil 值，
循环容器也受此限制；typed nil 输出 `null`。自定义 `json.Marshaler` 负责自身内容的序列化，
同时实现 error 时优先输出错误文本。`[]byte` 沿用标准库的 base64 编码。

仓库基准覆盖普通字段、嵌套容器及嵌套 error，并报告耗时与分配：

```sh
go test -run '^$' -bench '^BenchmarkJsonFormat$' -benchmem -count=3
```

其中 `DirectMarshal` 仅作为无 error 输入的标准库参照，不提供错误转换或深度限制。
基准只测格式化，不代表文件、控制台或网络输出的实际吞吐。

## 并发与兼容

core 支持并发日志、调整等级、设置预处理、增删输出器与替换默认 logger。
`RemovePrinter` 只解除注册，不关闭输出器；已经开始的分发可能继续。
`L.Close` 先拒绝新日志，等待在途分发，再关闭输出器。禁止在自己的回调中重入 Close。
自定义 Printer、格式器、mapper 与预处理函数必须并发安全；Builder 仅用于启动阶段构建。
`CompatibleLogger` 可作为 io.Writer、Printf 或字符串等级接口注入其他库。

v0.3.0 相对 v0.2.0.250209 的 zap API 是破坏性重构。本轮通用化还移除了
`Config.Name/Region`、默认 event logger、`Event/AddEventPrinter`、可变 context 的 `Bind/AddFields` 和 `ErrorFields`：
分别改用显式 `Fields/With`、独立数据 core、标准派生 context 和 `field.Error`。
`NewEventPrinter` 改名为 `NewDataPrinter`，日志目录直接采用 `LogFilePath`。

## 维护与验证

通用仓库约定见 [AGENTS.md](AGENTS.md)，本地编辑器和生成文件由 `.gitignore` 排除。

```sh
go test -race -count=1 ./...
go test -tags release -count=1 ./...
go vet ./...
```

测试覆盖字段与 context 隔离、显式映射、业务字段透明传递、并发配置与关闭、文件故障与队列策略。
真实 Sentry 服务和实际应用接入需要在对应环境联调。

## 社区与许可

- [获取帮助](SUPPORT.md) · [报告问题](https://github.com/NumberMan1/log/issues)
- [贡献指南](CONTRIBUTING.md) · [行为准则](CODE_OF_CONDUCT.md)
- [安全政策](SECURITY.md) · [维护者与发布流程](MAINTAINERS.md)
- [变更记录](CHANGELOG.md) · [第三方依赖说明](THIRD_PARTY_NOTICES.md)

项目原创内容采用 [MIT 许可证](LICENSE)，第三方依赖保留各自的版权与许可。
