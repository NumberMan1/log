# 贡献指南

欢迎提交缺陷修复、文档、示例、测试和性能改进。交流请遵守 [行为准则](CODE_OF_CONDUCT.md)；
使用问题见 [SUPPORT.md](SUPPORT.md)，漏洞见 [SECURITY.md](SECURITY.md)。

## 开始之前

先搜索已有 Issue。新增输出器、公开接口变更、关闭语义变更或依赖调整，
请先说明具体场景、兼容性及方案。Bug 报告请附上版本、Go 版本、操作系统、
相关配置和脱敏后的最小复现。中文和英文均可。

## 本地开发

要求 Go 1.25.0 或以上版本，模块路径为 `github.com/NumberMan1/log`，
根包名为 `serverlog`。代码与测试约定见 [AGENTS.md](AGENTS.md)。

在仓库根目录运行：

```sh
go mod download
go test -race -count=1 ./...
go test -tags release -count=1 ./...
go vet ./...
git diff --check
```

竞态检测需要目标平台支持并具有可用的 C 编译器。受限环境可以将 GOCACHE 指向可写临时目录。
Sentry 测试使用测试 transport，无需真实 DSN 或访问线上服务。

## 修改约定

- 本库保持通用，业务字段与框架适配由调用方显式提供。
- 使用 `gofmt` 格式化修改的 Go 文件，并为导出接口与配置字段补充准确的 Go 文档注释。
- 修改接口、输出语义、并发或生命周期时，同步更新 README、可运行示例及行为回归测试。
- 特别检查字段隔离、原始 error 保留、关闭时等待在途操作、队列满与后台 I/O 错误的可观察性。
- 自定义 Printer、格式器、mapper 与预处理函数必须满足文档中的并发要求。
- 性能调整附上可重复的基准结果；格式化基准不能替代实际输出吞吐测试。
- 依赖变动运行 `go mod tidy` 并提交 `go.mod` 和 `go.sum`，不夹带无关升级。
- 保留第三方许可，不提交凭据、业务日志、真实用户数据、IDE 配置或生成物。
- 需要让用户了解的变更记录到 [CHANGELOG.md](CHANGELOG.md)，破坏性变更说明迁移方式。

## Pull Request

从 fork 创建分支，一次 PR 聚焦一个问题。描述最终行为、解决的具体问题、兼容性与验证结果；
如未运行检查或存在失败，请明确说明。维护者评审通过后合并。

提交贡献表示你有权提供这些内容，并同意原创贡献采用项目 [MIT 许可证](LICENSE)；
第三方内容保留其自身许可，参见 [第三方说明](THIRD_PARTY_NOTICES.md)。
