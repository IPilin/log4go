package log4go

import (
	"fmt"
	"strings"
	"time"
)

type AppenderType string

const (
	AppenderTypeConsole AppenderType = "console"
	AppenderTypeFile    AppenderType = "file"
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

type Appender interface {
	Init() error
	Key() string
	Write(*Record) error
	Close() error
	Copy(*Appender) error
}

type Record struct {
	Time        time.Time
	Level       string
	PackageName string
	Data        string
}

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

	*l = level
	return nil
}
