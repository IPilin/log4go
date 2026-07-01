package log4go

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type ConsoleAppender struct {
	Target Target
	Format OutputFormat
	writer io.Writer
	mu     sync.Mutex
}

func (a *ConsoleAppender) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch a.Target {
	case TargetStdout:
		a.writer = os.Stdout
	case TargetStderr:
		a.writer = os.Stderr
	default:
		return fmt.Errorf("wrong target for ConsoleAppender: %q", a.Target)
	}
	return nil
}

func (a *ConsoleAppender) Key() string {
	return string(a.Target) + string(a.Format)
}

func writeDefault(level string, packageName string) *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	appendFormatTime(buf, time.Now())
	buf.WriteByte(' ')
	buf.WriteString(level)
	buf.WriteByte(' ')
	if len(packageName) != 0 {
		buf.WriteByte('[')
		buf.WriteString(packageName)
		buf.WriteString("] ")
	}
	return buf
}

func (a *ConsoleAppender) Write(r *Record) error {
	if a.writer == nil {
		return fmt.Errorf("no writer for ConsoleAppender{target: %q}", a.Target)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	buf := writeDefault(r.Level, r.PackageName)
	defer bufferPool.Put(buf)

	buf.WriteString(r.Data)
	buf.WriteByte('\n')
	_, err := a.writer.Write(buf.Bytes())
	return err
}

func (a *ConsoleAppender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return nil
}

func (a *ConsoleAppender) Copy(oldA *Appender) error {
	return nil
}
