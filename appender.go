package log4go

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type OutputFormat string

const (
	FormatText OutputFormat = "Text"
	FormatJson OutputFormat = "Json"
)

type Target string

const (
	TargetStdout Target = "stdout"
	TargetStderr Target = "stderr"
	TargetFile   Target = "file"
)

type Appender struct {
	Target Target       `yaml:"target"`
	Path   string       `yaml:"path"`
	Format OutputFormat `yaml:"format"`
	writer io.Writer
	file   *os.File
	mu     sync.Mutex
}

func (a *Appender) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch a.Target {
	case TargetStdout:
		a.writer = os.Stdout
	case TargetStderr:
		a.writer = os.Stderr
	case TargetFile:
		file, err := os.OpenFile(a.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		a.file = file
		a.writer = file
	}

	return nil
}

func (a *Appender) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil && a.writer == nil {
		return 0, fmt.Errorf("no writer for appender{target: %q, path: %q}", a.Target, a.Path)
	}

	return a.writer.Write(p)
}

func (a *Appender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		return nil
	}

	return a.file.Close()
}

type MultiAppender struct {
	Appenders []*Appender
	mu        sync.Mutex
}

func NewMultiAppender(v ...*Appender) *MultiAppender {
	ma := &MultiAppender{
		Appenders: v,
	}
	ma.Init()
	return ma
}

func (m *MultiAppender) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs error
	for _, a := range m.Appenders {
		err := a.Init()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (m *MultiAppender) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs error
	for _, a := range m.Appenders {
		_, err := a.Write(p)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return len(p), errs
}

func (m *MultiAppender) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs error
	for _, a := range m.Appenders {
		err := a.Close()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}
