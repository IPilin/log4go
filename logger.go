package log4go

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type appLogger struct {
	level     atomic.Int32
	ma        atomic.Pointer[MultiAppender]
	logTunnel chan []byte
	errTunnel chan error
	stop      chan struct{}
}

func (a *appLogger) Write(b []byte) (int, error) {
	select {
	case a.logTunnel <- b:
		return len(b), nil
	default:
		return 0, fmt.Errorf("log tunnel overflow, message dropped: %q", string(b))
	}
}

func Close() error {
	err := instance.ma.Load().Close()
	close(instance.logTunnel)
	close(instance.errTunnel)
	<-instance.stop
	<-instance.stop
	return err
}

type logMap struct {
	mu   sync.RWMutex
	logs map[string](*Log)
}

func (m *logMap) get(key string) *Log {
	m.mu.RLock()
	log, ok := m.logs[key]
	m.mu.RUnlock()

	if ok {
		return log
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if log, ok := m.logs[key]; ok {
		return log
	}
	log = &Log{packageName: key}

	m.logs[key] = log
	return log
}

var (
	instance = func() *appLogger {
		a := &appLogger{}

		ma, _ := NewMultiAppender(&Appender{
			Target: TargetStdout,
			Format: FormatText,
		})
		ma.Init()

		a.ma.Store(ma)
		a.level.Store(int32(InfoLevel))

		a.logTunnel = make(chan []byte, 1024)
		a.errTunnel = make(chan error, 100)
		a.stop = make(chan struct{}, 2)

		go bootstrapWriter(a)
		go bootstrapError(a)

		return a
	}()
	logs logMap = logMap{
		logs: make(map[string](*Log)),
	}
	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

func Init(config *LogConfig) error {
	config, err := initConfig(config)
	if err != nil {
		return err
	}

	ma, err := NewMultiAppender(config.Outputs...)
	if err != nil {
		return err
	}

	var errs error
	var newAppenders = make(map[string](*Appender))
	for key, val := range ma.Appenders {
		if oldApp, ok := instance.ma.Load().Appenders[key]; ok {
			val.file = oldApp.file
			val.writer = oldApp.writer
			continue
		}

		if err := val.Init(); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		newAppenders[key] = val
	}

	if errs != nil {
		for _, val := range newAppenders {
			val.Close()
		}
		return errs
	}

	oldMa := instance.ma.Load()

	instance.ma.Store(ma)
	instance.level.Store(int32(config.Level))

	for key, val := range oldMa.Appenders {
		if _, ok := ma.Appenders[key]; ok {
			continue
		}
		val.Close()
	}

	return nil
}

func bootstrapWriter(app *appLogger) {
	for b := range app.logTunnel {
		_, err := app.ma.Load().Write(b)
		if err != nil {
			select {
			case app.errTunnel <- err:
			default:
				fmt.Fprintf(os.Stderr, "log4go: error tunnel overflow, message dropped: %q", err)
			}
		}
	}
	app.stop <- struct{}{}
}

func bootstrapError(app *appLogger) {
	for err := range app.errTunnel {
		fmt.Fprintf(os.Stderr, "log4go: Error - %q", err)
	}
	app.stop <- struct{}{}
}

type Log struct {
	packageName string
}

func GetLog(packageName string) *Log {
	return logs.get(packageName)
}

func (l *Log) Trace(v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > TraceLevel {
		return 0, nil
	}

	return writeLogs("TRACE", l.packageName, v...)
}

func (l *Log) Tracef(format string, v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > TraceLevel {
		return 0, nil
	}

	return writeLogsf("TRACE", l.packageName, format, v...)
}

func (l *Log) Debug(v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > DebugLevel {
		return 0, nil
	}

	return writeLogs("DEBUG", l.packageName, v...)
}

func (l *Log) Debugf(format string, v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > DebugLevel {
		return 0, nil
	}

	return writeLogsf("DEBUG", l.packageName, format, v...)
}

func (l *Log) Info(v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > InfoLevel {
		return 0, nil
	}

	return writeLogs("INFO", l.packageName, v...)
}

func (l *Log) Infof(format string, v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > InfoLevel {
		return 0, nil
	}

	return writeLogsf("INFO", l.packageName, format, v...)
}

func (l *Log) Warn(v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > WarnLevel {
		return 0, nil
	}

	return writeLogs("WARN", l.packageName, v...)
}

func (l *Log) Warnf(format string, v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > WarnLevel {
		return 0, nil
	}

	return writeLogsf("WARN", l.packageName, format, v...)
}

func (l *Log) Error(v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > ErrorLevel {
		return 0, nil
	}

	return writeLogs("ERROR", l.packageName, v...)
}

func (l *Log) Errorf(format string, v ...any) (int, error) {
	if LogLevel(instance.level.Load()) > ErrorLevel {
		return 0, nil
	}

	return writeLogsf("ERROR", l.packageName, format, v...)
}

func writeLogs(level string, packageName string, v ...any) (int, error) {
	buf := writeDefault(level, packageName)
	defer bufferPool.Put(buf)

	fmt.Fprint(buf, v...)
	buf.WriteByte('\n')

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	return instance.Write(data)
}

func writeLogsf(level string, packageName string, format string, v ...any) (int, error) {
	buf := writeDefault(level, packageName)
	defer bufferPool.Put(buf)

	fmt.Fprintf(buf, format, v...)
	buf.WriteByte('\n')

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	return instance.Write(data)
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
