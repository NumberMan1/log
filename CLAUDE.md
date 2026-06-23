# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 构建和测试

```bash
# 构建
go build ./...

# 运行全部测试
go test -count=1 ./...

# 运行全部测试（含竞态检测）
go test -race -count=1 ./...

# 运行单个包的测试
go test -v -count=1 ./fileutil/...

# 运行单个测试用例
go test -v -count=1 -run TestLogger_Debug .

# 覆盖率
go test -coverprofile=/tmp/cover.out ./... && go tool cover -func=/tmp/cover.out
```

go.mod 中声明 `go 1.19`，但实际工作区使用更新的 Go 工具链（1.25.x）。依赖通过 `go mod tidy` 管理。

## 架构

`serverlog` 包（导入路径 `github.com/NumberMan1/log`）是一个基于 **Printer + LevelSwitch** 管道模型的分级结构化日志库。

### 核心概念

- **`L`**（`log.go`）—— 日志记录器。内部持有 `map[Printer]*LevelSwitch`、`maxLevel`、可选的 `preHandler`。`Log(level, msg, data)` 遍历所有 printer，检查 level switch 后输出。Panic/Fatal 级别分别触发 panic/exit（exit 可通过 `SetExitFunc` 覆盖，方便测试）。
- **`Printer`** 接口（`contract.go`）—— `Log(level, msg, data)` + `Close()`。具体实现：`ConsolePrinter`、`FilePrinter`、`SentryPrinter`。
- **`LevelSwitch`**（`level_switcher.go`）—— 基于位掩码的级别开关。方法：`EnableLevels`、`EnableSingleLevel`、`DisableSingleLevel`、`EnableLevelLte`（启用该级别及以下所有级别）、`Reset`、`IsEnabled`。通过 `atomic.Uint32` 保证并发安全。
- **`Level`**（`contract.go`）—— `uint32` iota 枚举：`PanicLevel`(0) → `TraceLevel`(6)。级别排序：Panic < Fatal < Error < Warn < Info < Debug < Trace（数值越大严重程度越低）。

### 数据流

```
Log(level, msg, data)
  → maxLevel 关卡（level > maxLevel 时跳过）
  → ErrorMsgPreHandler（将 error 类型值转为字符串）
  → 自定义 preHandler（如果有）
  → 遍历 {Printer, LevelSwitch}:
      若 LevelSwitch.IsEnabled(level) → printer.Log(level, msg, data)
        → printer.preHandler(data)   （每个 printer 有自己的 preHandler）
        → printer.logFormat.Format() （每个 printer 有自己的 format）
  → PanicLevel/FatalLevel 时触发 panic/exit
```

### 关键模块

| 文件 | 职责 |
|---|---|
| `log.go` | Logger 结构体、全局单例（`gLogger`、`gEventLogger`）、exit 覆盖机制 |
| `contract.go` | `Printer`、`LogFormat`、`PreHandler` 接口；`Level` 类型 |
| `level_switcher.go` | 基于位掩码的级别过滤 |
| `builder.go` | 流式 Builder API，用于构建预配置的 logger |
| `compatible.go` | `CompatibleLogger`——将 `L` 适配为 `io.Writer` + `Printf`，供第三方库注入使用 |

### PreHandler（可组合的数据转换器）

- `AddTimePreHandler` —— data 非 nil 时添加 `"Time"`（unix 时间戳）
- `ErrorMsgPreHandler` —— 由 `L.Log` 自动调用；将 `error`/`err` 键从 `error` 类型转为 `.Error()` 字符串
- `StackWrapPreHandler` / `StacksWrapPreHandler` —— 用 `Time`、`LogType`、`Message`、`Stack`（单帧或全部帧）包裹 data
- `ConcatPreHandler` —— 将多个 PreHandler 串联执行

### Printer

- **`ConsolePrinter`** —— 写入 `os.Stdout`。可配置 `LogFormat`（默认 `TextFormat`）、`PreHandler`、`SkipFrame`。
- **`FilePrinter`** —— 异步文件写入，带缓冲 I/O、定时刷新、通过 `fileutil.Switcher` 实现文件轮转。默认格式：`JsonFormat`。
- **`SentryPrinter`** —— 通过 `sentry-go` 上报 Sentry。使用 `Contexts["extra"]` 存储额外数据（API 在 v0.47 版本中移除了 `Event.Extra` 字段）。
- 工厂函数：`NewJsonlogPrinter`、`NewEventPrinter` 生成预配置的 `FilePrinter` 实例。

### Format

- `TextFormat` —— `[WARN]\t<RFC3339>\t<文件:行号>\t<消息> <按key排序的键值对>`
- `JsonFormat` —— `json.Marshal(data)`，data 为 nil 时返回 `""`
- `OnlyMessageFormat` —— 原样返回 msg

### fileutil 子包

- `File` —— 封装 `*os.File` + `*bufio.Writer`，通过 mutex 保护 `Write`、`Flush`、`Close`。
- `Switcher` 接口 —— `FirstSwitchTime()`、`NextSwitchTime()`、`DirAndFileName()`。两个实现：`HOURS_SWITCHER`（按小时）和 `DAY_SWITCHER`（按天）用于日志轮转。
- `FileExists` —— 基于 stat 的文件存在性检查。

### 双 Logger 设计

库维护两个独立的全局 logger：
- **`gLogger`** —— 用于应用日志（Debug、Info、Error、Fatal、Panic）。对应包级函数 `Debug()`、`Info()`、`Error()`、`Fatal()`、`Panic()`。
- **`gEventLogger`** —— 用于结构化事件数据。对应包级函数 `Event()`。事件日志与应用日志隔离，避免混入。

### Build tags

- `default_dev.go`（`//go:build !release`）—— `defaultMaxLevel = DebugLevel`
- `default_release.go`（`//go:build release`）—— `defaultMaxLevel = InfoLevel`

`default.go` 中的 init 会读取 `MAX_LOG_LEVEL` 环境变量（覆盖 build tag 默认值）并调用 `SetMaxLevel`。
