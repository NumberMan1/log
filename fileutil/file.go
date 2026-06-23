package fileutil

import (
	"bufio"
	"os"
	"sync"
)

type File struct {
	mutex   sync.Mutex
	f       *os.File
	w       *bufio.Writer
	changed bool
}

func NewFile(fullName string, writeBufferSize int) (*File, error) {
	f, err := os.OpenFile(fullName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriterSize(f, writeBufferSize)
	return &File{
		f: f,
		w: w,
	}, nil
}

func (file *File) Write(data string) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	if _, err := file.w.WriteString(data); err != nil {
		return err
	}

	file.changed = true
	return nil
}

func (file *File) WriteBytes(bytes []byte) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	if _, err := file.w.Write(bytes); err != nil {
		return err
	}

	file.changed = true
	return nil
}

func (file *File) Flush() error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

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

func (file *File) Close() error {
	if err := file.Flush(); err != nil {
		return err
	}
	return file.f.Close()
}
