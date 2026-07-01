package log4go

import (
	"fmt"
	"time"
)

type Log struct {
	packageName string
}

func (l *Log) Trace(v ...any) error {
	if LogLevel(instance.level.Load()) > TraceLevel {
		return nil
	}

	return writeLogs("TRACE", l.packageName, v...)
}

func (l *Log) Tracef(format string, v ...any) error {
	if LogLevel(instance.level.Load()) > TraceLevel {
		return nil
	}

	return writeLogsf("TRACE", l.packageName, format, v...)
}

func (l *Log) Debug(v ...any) error {
	if LogLevel(instance.level.Load()) > DebugLevel {
		return nil
	}

	return writeLogs("DEBUG", l.packageName, v...)
}

func (l *Log) Debugf(format string, v ...any) error {
	if LogLevel(instance.level.Load()) > DebugLevel {
		return nil
	}

	return writeLogsf("DEBUG", l.packageName, format, v...)
}

func (l *Log) Info(v ...any) error {
	if LogLevel(instance.level.Load()) > InfoLevel {
		return nil
	}

	return writeLogs("INFO", l.packageName, v...)
}

func (l *Log) Infof(format string, v ...any) error {
	if LogLevel(instance.level.Load()) > InfoLevel {
		return nil
	}

	return writeLogsf("INFO", l.packageName, format, v...)
}

func (l *Log) Warn(v ...any) error {
	if LogLevel(instance.level.Load()) > WarnLevel {
		return nil
	}

	return writeLogs("WARN", l.packageName, v...)
}

func (l *Log) Warnf(format string, v ...any) error {
	if LogLevel(instance.level.Load()) > WarnLevel {
		return nil
	}

	return writeLogsf("WARN", l.packageName, format, v...)
}

func (l *Log) Error(v ...any) error {
	if LogLevel(instance.level.Load()) > ErrorLevel {
		return nil
	}

	return writeLogs("ERROR", l.packageName, v...)
}

func (l *Log) Errorf(format string, v ...any) error {
	if LogLevel(instance.level.Load()) > ErrorLevel {
		return nil
	}

	return writeLogsf("ERROR", l.packageName, format, v...)
}

func writeLogs(level string, packageName string, v ...any) error {
	r := Record{
		Time:        time.Now(),
		Level:       level,
		PackageName: packageName,
		Data:        fmt.Sprint(v...),
	}

	return instance.Write(&r)
}

func writeLogsf(level string, packageName string, format string, v ...any) error {
	r := Record{
		Time:        time.Now(),
		Level:       level,
		PackageName: packageName,
		Data:        fmt.Sprintf(format, v...),
	}

	return instance.Write(&r)
}
