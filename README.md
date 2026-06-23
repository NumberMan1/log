# 介绍

实现系统信息日志模块，根据不同的日志接口分别输出日志


## 用法

serverlog模块可以全局的方式使用，使用时需要在程序启动的时候一次或多次调用`serverlog.AddPrinter(printer)` 来初始化日志的打印器，程序退出时调用`serverlog.Close()`来关闭全局日志系统

- FilePrinter -- 日志输出到文件
- ConsolePrinter -- 日志输出到终端
- SentryPrinter -- 日志输出到sentry，`Init`时传入sentry_dsn即可开始使用，所需要的环境变量：设置sentry环境 `SENTRY_ENVIRONMENT`, 设置服务器名称: `SERVER_NAME`, 设置sentry release: `SENTRY_RELEASE`

记录日志的接口分为以下几种：

1. Debug() -- 用于记录调试信息，通常是开发时才关心的一些调试信息
2. Info() -- 用于记录信息，通常是一些跟系统运行情况相关的数据，比如程序启动时间
3. Error() -- 用于记录错误，通常是一些业务失败或操作失败，比如非法请求或请求处理失败
4. Fatal() -- 用于记录异常，通常是一些非预期的错误，比如空指针
5. Event() -- event并不属于Logger,而是封装了一个gEventLogger来打印事件日志，使用前需要调用 AddEventPrinter 添加事件打印器

默认的 `Debug`和`Info`默认输出到标准输出；`Event`以JSON格式记录日志文件，按小时自动切换；`Error`和`Fatal`上报到`Sentry`后台

#### 快速开始
```go
// sentryDsn不用可以填"""
serverlog.Init(logPath, sentryDsn)
```

- 常用的日志打印是打印到控制台和文件，`InitDefault`提供了这样的支持，如果需要输出到Sentry需要添加 `SentryPrinter`。
- 通过InitDefault初始化，可以通过配置环境变量 `ALL_LOG_CONSOLE=true` 使所有的日志都打印到控制台上。
- release标签下默认的最大日志级别是 `INFO`，非release标签下最大的日志级别是 `DEBUG`，可以启动程序前设置环境变量 `MAX_LOG_LEVEL` 为 **trace/debug/info/warn/error/fatal/panic** 中的一个来修改最大日志级别
- 如果添加 `SentryPrinter`，必须确保执行过 `InitSentry(dsn, environment)`, `serverlog.Init()` 已经有执行，不通过 `serverlog.Init()` 初始化则:
  ```go
  serverlog.InitSentry("https://ddfe141df54e49d88e90d94e633d398d@sentry.sausage.xd.com/2", "develop")
  serverlog.AddPrinter(serverlog.NewSentryPrinter(serverlog.WithExtraTagOption(func(tags map[string]string) {
		tags["sid"] = sid
	}))
  ```

#### 其他方法

- 使用Builder模式构建logger:
  ```go
  logger := NewBuilder().AddConsolePrinter(InfoLevel, serverlog.ConsolePrinterConfig{}).AddJsonlogPrinter(InfoLevel, "./log/jsonlog").Logger()
  ```

- 添加打印器:
  ```go
  # LevelSwitch 配置想要在哪个等级中打印使用
  consolePrinter := serverlog.NewConsolePrinter()
  serverlog.AddPrinter(NewLevelSwitch().EnableLevels(serverlog.DebugLevel, serverlog.InfoLevel), consolePrinter)
  serverlog.AddEventPrinter(consolePrinter)
  ```
- 移除打印器
  ```go
  serverlog.RemovePrinter(consolePrinter)
  ```
- 修改打印器的打印级别
  ```go
  ls := serverlog.GetLevelSwitcherByPrinter(consolePrinter)
  ls.EnableSingleLevel(serverlog.EventLevel)
  ```

- 修改日志最大打印级别
  ```go
  serverlog.SetMaxLevel(serverlog.DebugLevel)
  logger.SetMaxLevel(serverlog.DebugLevel)
  ```

- 启动阶段日志
  ```go
  // 自定义打印器
  // serverlog.SetBootInfoPrinter(func(v ...any){fmt.Println(v...)})
  serverlog.BootInfo("Bootstrap... step 1.")
  ```

----

#### 单独logger打印
serverlog.RemovePrinter 这类包级别的函数使用的是默认的全局logger，也可以通过 `NewLogger` 来生成新的logger
  ```go
	logger := NewLogger()
	logger.AddPrinter(NewLevelSwitch().EnableLevels(DebugLevel, InfoLevel), NewConsolePrinter())
	ep, err := NewEventPrinter(dir)
	if err != nil {
		t.Fatal(err)
	}
	logger.AddPrinter(NewLevelSwitch().EnableSingleLevel(EventLevel).EnableLevelLte(ErrorLevel), ep)
	logger.AddPrinter(NewLevelSwitch().EnableLevelLte(ErrorLevel), NewSentryPrinter())
	defer logger.Close()

	logger.Info("HelloLog")
  ```

----

#### 暴露 `CompatibleLogger` 方便做为其他库的组件
> 有些地方在使用 serverlog 打印的时候并不关心等级，缺少等级的情况下将日志以 InfoLevel 打印。
##### 获取一个可以给第三方用的 serverlog:
```go
		logger := serverlog.NewLogger()
		logger.AddPrinter(serverlog.NewLevelSwitch().EnableLevels(InfoLevel), serverlog.NewConsolePrinterDefault())
		filePrinter, err := serverlog.NewFilePrinter(serverlog.FilePrinterConfig{
			Dir:       "/data/log/",
			LogFormat: serverlog.NewOnlyMessageFormat(), // 由于writer的方式没有 level和M 参数，所以需要格式转换器只打 message
		})
		if err != nil {
			panic(err)
		}
		logger.AddPrinter(serverlog.NewLevelSwitch().EnableLevels(InfoLevel), filePrinter)
		compatibleLogger := serverlog.NewCompatibleLogger(logger)
```

##### 第三方兼容serverlog的方式:
1. 允许传入 `io.Writer` 并在需要打印日志的地方调用 `Write` 方法
  示例:
  ```go
  var out io.Writer

  func SetOut(injectOut io.Writer) {
    out = injectOut
  }

  func main() {
    l := serverlog.NewCompatibleLogger(serverlog.GetDefaultLogger())
    SetLogger(l) 
  }

  ```

2. 自定义 `interface{ Log(levelStr string, message string, data map[string]interface{}) }`，接受该 interface 的流入
  示例:
  ```go
  type logger interface {
    Log(levelStr string, message string, data map[string]interface{})
  }

  var l logger

  func SetLogger(injectLog logger) {
    l = injectLog
  }

  func main() {
    l := serverlog.NewCompatibleLogger(serverlog.GetDefaultLogger())
    SetLogger(l)
  }
  ```

----

在很多时候我们需要用到key-value的形式来记录日志数据，所以serverlog模块内置了一个M类型，用来做这件事情。

M类型是一个用字符串做key，存放`interface{}`类型的map，所以它可以存放任何可以参与JSON序列化的数据。

以下是一些用法示例：

```go
serverlog.Info("Hello %s", "World")

serverlog.Event(serverlog.M{
    "error": error.Error(),
    "pid": pid,
    "group_id": groupID,
})

serverlog.Error("get link failed", serverlog.M{
    "sid":     env.ServerId(),
    "error":   fmt.Sprintf("%s", err),
    "id":      id,
    "network": node.Network,
    "address": node.Address,
})

serverlog.Fatal("DoTransaction failed", serverlog.M{
    "sid":   env.ServerId(),
    "error": fmt.Sprintf("%s", e),
    "pid":   pid,
    "msg":   fmt.Sprintf("%v", errInfo),
})

```
