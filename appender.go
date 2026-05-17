package log4go

import (
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

func (a *ConsoleAppender) Write(r *Record) error {
	if a.writer == nil {
		return fmt.Errorf("no writer for ConsoleAppender{target: %q}", a.Target)
	}
	return a.writer.Write(r)
}

func (a *ConsoleAppender) Close() error {
	return nil
}

func (a *ConsoleAppender) Copy(oldA *Appender) error {
	return nil
}
