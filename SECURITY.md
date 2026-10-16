# 安全政策

## 维护范围

安全修复优先落在默认分支，并按需发布新的 tag。旧版本不承诺回补；
请使用包含修复的版本。已发布版本见 [Releases](https://github.com/NumberMan1/log/releases)
和 [Tags](https://github.com/NumberMan1/log/tags)。

## 私下报告漏洞

请优先使用 [GitHub 私密漏洞报告](https://github.com/NumberMan1/log/security/advisories/new)。
此入口需要仓库管理员在 GitHub 设置中启用 private vulnerability reporting。

若入口不可用，请通过 [维护者](MAINTAINERS.md) GitHub 主页公开提供的私下联系渠道沟通；
若没有可用渠道，只提交不含漏洞细节的 Issue，请求开启私密报告入口。
不要在公开 Issue、PR 或日志中提交利用代码、访问凭据或真实个人数据。

报告请尽量包含：

- 受影响的版本或提交、Go 版本、操作系统及相关配置。
- 漏洞类型、影响范围与触发条件。
- 使用虚构或脱敏数据的最小复现步骤。
- 已知缓解方式、修复建议，以及是否希望在公告中署名。

维护者会评估影响、协商修复与披露方式；没有固定响应时间或修复时限承诺。
在修复或协调披露前，请保持漏洞细节私密。
