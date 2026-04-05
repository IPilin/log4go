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
	//TODO: make Json formating
	//FormatJson OutputFormat = "Json"
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

// Init initializes the appender's underlying writer based on the Target.
// It opens the file if Target is TargetFile, otherwise sets os.Stdout or os.Stderr.
// Returns an error if the file cannot be opened.
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

// Write bytes to Target
func (a *Appender) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writer == nil {
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
	Appenders map[string](*Appender)
	mu        sync.RWMutex
}

func NewMultiAppender(v ...*Appender) (*MultiAppender, error) {
	ma := &MultiAppender{
		Appenders: func(v ...*Appender) map[string](*Appender) {
			a := make(map[string](*Appender))
			for _, val := range v {
				if len(val.Path) == 0 {
					a[string(val.Target)] = val
				} else {
					a[val.Path] = val
				}
			}
			return a
		}(v...),
	}
	if err := ma.Init(); err != nil {
		return nil, err
	}
	return ma, nil
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

func (m *MultiAppender) Update(newMa *MultiAppender) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, val := range newMa.Appenders {
		if _, ok := m.Appenders[key]; ok {
			continue
		}
		if err := val.Init(); err != nil {
			newMa.Close()
			return err
		}
		m.Appenders[key] = val
	}

	var errs error
	for key, val := range m.Appenders {
		if _, ok := newMa.Appenders[key]; ok {
			continue
		}
		if err := val.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
		delete(m.Appenders, key)
	}
	return errs
}

func (m *MultiAppender) Write(p []byte) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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
