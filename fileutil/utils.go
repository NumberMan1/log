package fileutil

import "os"

// 文件是否存在
func FileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
