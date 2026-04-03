package log4go

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/IPilin/log4go/utils"
)

type appLogger struct {
	level LogLevel
	ma    *MultiAppender
}

var (
	instance   *appLogger
	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

func Init(config *LogConfig) (err error) {
	config, err = initConfig(config)
	if err != nil {
		return err
	}

	ma := NewMultiAppender(config.Outputs...)

	instance = &appLogger{
		level: config.Level,
		ma:    ma,
	}

	return nil
}

type Log struct {
	PackageName string
}

func (l *Log) Trace(v ...any) {
	if instance.level > TraceLevel {
		return
	}

	writeLogs("TRACE", l.PackageName, v...)
}

func (l *Log) Tracef(format string, v ...any) {
	if instance.level > TraceLevel {
		return
	}

	writeLogsf("TRACE", l.PackageName, format, v...)
}

func (l *Log) Debug(v ...any) {
	if instance.level > DebugLevel {
		return
	}

	writeLogs("DEBUG", l.PackageName, v...)
}

func (l *Log) Debugf(format string, v ...any) {
	if instance.level > DebugLevel {
		return
	}

	writeLogsf("DEBUG", l.PackageName, format, v...)
}

func (l *Log) Info(v ...any) {
	if instance.level > InfoLevel {
		return
	}

	writeLogs("INFO", l.PackageName, v...)
}

func (l *Log) Infof(format string, v ...any) {
	if instance.level > InfoLevel {
		return
	}

	writeLogsf("INFO", l.PackageName, format, v...)
}

func (l *Log) Warn(v ...any) {
	if instance.level > WarnLevel {
		return
	}

	writeLogs("WARN", l.PackageName, v...)
}

func (l *Log) Warnf(format string, v ...any) {
	if instance.level > WarnLevel {
		return
	}

	writeLogsf("WARN", l.PackageName, format, v...)
}

func (l *Log) Error(v ...any) {
	if instance.level > ErrorLevel {
		return
	}

	writeLogs("ERROR", l.PackageName, v...)
}

func (l *Log) Errorf(format string, v ...any) {
	if instance.level > ErrorLevel {
		return
	}

	writeLogsf("ERROR", l.PackageName, format, v...)
}

func writeLogs(level string, packageName string, v ...any) {
	buf := writeDefault(level, packageName)
	defer bufferPool.Put(buf)

	fmt.Fprint(buf, v...)
	buf.WriteByte('\n')

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	go instance.ma.Write(data)
}

func writeLogsf(level string, packageName string, format string, v ...any) {
	buf := writeDefault(level, packageName)
	defer bufferPool.Put(buf)

	fmt.Fprintf(buf, format, v...)
	buf.WriteByte('\n')

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	go instance.ma.Write(data)
}

func writeDefault(level string, packageName string) *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	utils.AppendFormatTime(buf, time.Now())
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
