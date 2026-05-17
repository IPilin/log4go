package log4go

import (
	"fmt"
	"strings"
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

	*l = level
	return nil
}
