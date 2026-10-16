# 维护者与发布

维护者：[@NumberMan1](https://github.com/NumberMan1)。

维护者负责 Issue 分流、代码评审、合并与版本发布。
实质性接口变更先在 Issue 或 PR 中讨论；贡献者通过 PR 参与。
评审责任配置见 [.github/CODEOWNERS](.github/CODEOWNERS)。

## 发布流程

1. 确认目标提交的 CI 通过，并完成 [贡献指南](CONTRIBUTING.md) 中的本地检查。
2. 检查 README、示例、最低 Go 版本、依赖与第三方许可说明是否一致。
3. 将 [CHANGELOG.md](CHANGELOG.md) 的未发布条目整理为实际版本和发布日期，说明不兼容变更与迁移方式。
4. 按语义化版本选择新的 `vX.Y.Z` tag；v0 阶段也应明确记录破坏性变更。
   若未来发布 v2 及以上主版本，须同步处理 Go module 的主版本路径。
5. 将 tag 推送到 GitHub，在对应 Release 中说明变更、迁移步骤与验证范围。
   已发布 tag 不覆盖，修正内容使用新版本。

## GitHub 管理设置

以下配置需要仓库管理员在 GitHub 中设置，仓库内文件不会自动启用：

- 允许 GitHub Actions 和 Dependabot 运行。
- 启用 private vulnerability reporting，使 [安全报告入口](SECURITY.md) 可用。
- 为默认分支设置 PR 评审与必要 CI 检查；是否强制 CODEOWNERS 审批由维护者决定。
