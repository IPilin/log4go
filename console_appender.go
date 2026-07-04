package log4go

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
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

func (a *ConsoleAppender) writeDefault(r *Record) *bytes.Buffer {
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

func (a *ConsoleAppender) Write(r *Record) error {
	if a.writer == nil {
		return fmt.Errorf("no writer for ConsoleAppender{target: %q}", a.Target)
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

func (a *ConsoleAppender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return nil
}

func (a *ConsoleAppender) Copy(oldA *Appender) error {
	return nil
}
