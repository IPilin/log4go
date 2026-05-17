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

type Record struct {
	Time        time.Time
	Level       string
	PackageName string
	Format      string
	Data        []any
}

type Appender interface {
	Init() error
	Key() string
	Write(*Record) error
	Close() error
	Copy(*Appender) error
}

type appLogger struct {
	level     atomic.Int32
	ma        atomic.Pointer[MultiAppender]
	logTunnel chan *Record
	errTunnel chan error
	stop      chan struct{}
}

func (a *appLogger) Write(r *Record) error {
	select {
	case a.logTunnel <- r:
		return nil
	default:
		return fmt.Errorf("log tunnel overflow, message dropped: %q", r)
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

		ma, _ := NewMultiAppender(&AppenderConfig{
			Type:   AppenderTypeConsole,
			Target: TargetStdout,
			Format: FormatText,
		})

		ma.Init()

		a.ma.Store(ma)
		a.level.Store(int32(InfoLevel))

		a.logTunnel = make(chan *Record, 1024)
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
	var newAppenders = make(map[string](Appender))
	for key, val := range ma.Appenders {
		if oldApp, ok := instance.ma.Load().Appenders[key]; ok {
			ma.Appenders[key] = oldApp
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
	for r := range app.logTunnel {
		err := app.ma.Load().Write(r)
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
	// buf := writeDefault(level, packageName)
	// defer bufferPool.Put(buf)

	// fmt.Fprint(buf, v...)
	// buf.WriteByte('\n')

	// data := make([]byte, buf.Len())
	// copy(data, buf.Bytes())

	r := Record{
		Time:        time.Now(),
		Level:       level,
		PackageName: packageName,
		Data:        v,
	}

	return instance.Write(&r)
}

func writeLogsf(level string, packageName string, format string, v ...any) error {
	// buf := writeDefault(level, packageName)
	// defer bufferPool.Put(buf)

	// fmt.Fprintf(buf, format, v...)
	// buf.WriteByte('\n')

	// data := make([]byte, buf.Len())
	// copy(data, buf.Bytes())

	r := Record{
		Time:        time.Now(),
		Level:       level,
		PackageName: packageName,
		Format:      format,
		Data:        v,
	}

	return instance.Write(&r)
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
