// Package serverlog 提供与业务领域无关的结构化日志、等级路由及输出器。
//
// L 负责配置、分发与关闭，Logger 提供不可变字段 scope，context 入口使用标准 context.Context。
// core 在分发时复制顶层字段 map，并保留原始 error；嵌套 payload 需由调用方保持只读。
// Console、File 和 Sentry 是可组合输出，应用身份、用户映射和字段命名均由调用方显式决定。
//
// 创建独立 core 后应在生命周期结束时调用 Close；字段 scope 不单独关闭共享 core。
// 自定义格式器、处理器与输出器必须并发安全，不得在自己的回调中关闭持有它们的 core。
package serverlog
