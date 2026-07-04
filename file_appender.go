package log4go

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

type FileAppender struct {
	FilePath string
	Format   OutputFormat
	writer   io.Writer
	file     os.File
	mu       sync.Mutex
}

func (a *FileAppender) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.writer = file
	return nil
}

func (a *FileAppender) Key() string {
	return string(a.FilePath) + string(a.Format)
}

func (a *FileAppender) writeDefault(r *Record) *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	appendFormatTime(buf, r.Time)
	buf.WriteByte(' ')
	buf.WriteString(r.Level)
	buf.WriteByte(' ')
	if len(r.PackageName) != 0 {
		buf.WriteByte('[')
		buf.WriteString(r.PackageName)
		buf.WriteString("] ")
	}
	return buf
}

func (a *FileAppender) Write(r *Record) error {
	if a.writer == nil {
		return fmt.Errorf("no writer for FileAppender{target: %q}", a.FilePath)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	buf := a.writeDefault(r)
	defer bufferPool.Put(buf)

	buf.WriteString(r.Data)
	buf.WriteByte('\n')
	_, err := a.writer.Write(buf.Bytes())
	return err
}

func (a *FileAppender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return nil
}

func (a *FileAppender) Copy(oldA *Appender) error {
	return nil
}
