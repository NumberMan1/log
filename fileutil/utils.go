package fileutil

import "os"

// FileExists 检查路径是否可被 Stat；目录也返回 true，其他错误返回 false。
func FileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
