package log4go

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type LogLevel int

const (
	TraceLevel LogLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l *LogLevel) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var level LogLevel
	switch strings.ToUpper(s) {
	case "TRACE":
		level = TraceLevel
	case "DEBUG":
		level = DebugLevel
	case "INFO":
		level = InfoLevel
	case "WARN":
		level = WarnLevel
	case "ERROR":
		level = ErrorLevel
	default:
		level = InfoLevel
		fmt.Printf("log4go: unknown level %q, falling back to INFO\n", s)
	}

	l = &level
	return nil
}

type OutputFormat = string

const (
	Text OutputFormat = "Text"
	Json OutputFormat = "Json"
)

type Target = string

const (
	Stdout Target = "stdout"
	Stderr Target = "stderr"
	File   Target = "file"
)

type LogConfig struct {
	ConfigPath string
	HotReload  bool        `yaml:"hot_reload"`
	Level      LogLevel    `yaml:"level"`
	Outputs    []*Appender `yaml:"outputs"`
}

type Appender struct {
	Target Target `yaml:"target"`
	Path   string `yaml:"path"`
	Format string `yaml:"format"`
	writer io.Writer
	file   *os.File
	mu     sync.Mutex
}

func (a *Appender) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch a.Target {
	case Stdout:
		a.writer = os.Stdout
	case Stderr:
		a.writer = os.Stderr
	case File:
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
