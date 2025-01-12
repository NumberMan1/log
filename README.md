# Logger

> 基于``` go.uber.org/zap ```封装实现的日志库

## Quick Start

### 日志配置

目录：``` config/conf.yaml ```

```yaml
服务配置文件层级
...
...
...
logger:
  # 日志级别
  level: "debug"
  # 是否输出到文件
  fileout: true
  # 日志文件地址，当fileout=true时生效
  logDir: "~/data/log/campfire"
  # 是否输出到控制台
  stdout: true

```
