# 第三方依赖与许可说明

根目录 [LICENSE](LICENSE) 用于本项目的原创内容，不替代依赖的原始版权与许可证。

依赖版本以 [go.mod](go.mod) 和 [go.sum](go.sum) 为准。当前运行依赖包括：

| 模块 | 用途 | 上游 |
| --- | --- | --- |
| `github.com/getsentry/sentry-go` | Sentry 输出与事件传输 | [getsentry/sentry-go](https://github.com/getsentry/sentry-go) |
| `golang.org/x/sys` | 间接系统接口依赖 | [Go sys 仓库](https://go.googlesource.com/sys) |
| `golang.org/x/text` | 间接文本处理依赖 | [Go text 仓库](https://go.googlesource.com/text) |

使用对应版本模块目录内的 LICENSE、NOTICE 或其他许可文件作为权利与再分发条件的依据。
发布包含依赖源码的 vendor、压缩包或其他再分发物时，保留这些文件及版权声明。

添加或更新第三方代码时，同时记录来源、版本、修改范围和许可；
不要因为本项目使用 MIT，就覆盖第三方的原始许可。
