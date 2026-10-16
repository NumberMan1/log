package fileutil

import (
	"bufio"
	"errors"
	"os"
	"sync"
)

// File 是并发安全的带缓冲文件，串行化写入、Flush 与 Close。
// 构造后不要按值复制，Close 即使刷新失败也会关闭文件描述符。
type File struct {
	mutex    sync.Mutex
	f        *os.File
	w        *bufio.Writer
	changed  bool
	closed   bool
	closeErr error
}

// NewFile appends to an existing file instead of overwriting its beginning.
func NewFile(fullName string, writeBufferSize int) (*File, error) {
	return openFile(fullName, writeBufferSize, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
}

// NewExclusiveFile reserves a new name atomically, for concurrent loggers/processes.
func NewExclusiveFile(fullName string, writeBufferSize int) (*File, error) {
	return openFile(fullName, writeBufferSize, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
}

func openFile(name string, bufferSize, flags int) (*File, error) {
	f, err := os.OpenFile(name, flags, 0644)
	if err != nil {
		return nil, err
	}
	return &File{f: f, w: bufio.NewWriterSize(f, bufferSize)}, nil
}

// Write 向缓冲追加字符串，关闭后返回 os.ErrClosed；不保证立即落盘。
func (file *File) Write(data string) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if file.closed {
		return os.ErrClosed
	}
	n, err := file.w.WriteString(data)
	file.changed = file.changed || n > 0
	return err
}

// WriteBytes 向缓冲追加字节，关闭后返回 os.ErrClosed；不保留输入切片。
func (file *File) WriteBytes(data []byte) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if file.closed {
		return os.ErrClosed
	}
	n, err := file.w.Write(data)
	file.changed = file.changed || n > 0
	return err
}

// Flush 将已变更的缓冲写出并同步到存储，关闭后返回 os.ErrClosed。
func (file *File) Flush() error {
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if file.closed {
		return os.ErrClosed
	}
	return file.flush()
}

func (file *File) flush() error {
	if !file.changed {
		return nil
	}
	if err := file.w.Flush(); err != nil {
		return err
	}
	if err := file.f.Sync(); err != nil {
		return err
	}
	file.changed = false
	return nil
}

// Close 尝试刷新并始终关闭描述符；重复调用返回首次关闭结果。
func (file *File) Close() error {
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if file.closed {
		return file.closeErr
	}
	// Close the descriptor even when flushing fails.
	file.closeErr = errors.Join(file.flush(), file.f.Close())
	file.closed = true
	return file.closeErr
}
