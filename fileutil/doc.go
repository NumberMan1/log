// Package fileutil 提供并发安全的缓冲文件与可配置文件轮转策略。
//
// NewFile 追加已有文件，NewExclusiveFile 独占创建新文件。Write 仅写缓冲，
// Flush 将数据写出并同步存储；Close 即使遇到刷新故障也会释放文件描述符。
// 目录、文件命名和轮转时间由 Switcher 决定，不依赖业务身份或部署框架。
package fileutil
