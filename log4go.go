package log4go

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type appLogger struct {
	level     atomic.Int32
	closed    bool
	mu        sync.RWMutex
	ma        atomic.Pointer[MultiAppender]
	logTunnel chan *Record
	errTunnel chan error
	wg        sync.WaitGroup
}

func (a *appLogger) Write(r *Record) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed {
		return errors.New("logger already closed")
	}

	select {
	case a.logTunnel <- r:
		return nil
	default:
		return fmt.Errorf("log tunnel overflow, message dropped: %q", r)
	}
}

type logMap struct {
	mu   sync.RWMutex
	logs map[string]*Log
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
		_ = ma.Init()

		a.ma.Store(ma)
		a.level.Store(int32(InfoLevel))
		a.logTunnel = make(chan *Record, 1024)
		a.errTunnel = make(chan error, 100)

		a.wg.Add(2)
		go bootstrapWriter(a)
		go bootstrapError(a)

		return a
	}()
	logs = logMap{
		logs: make(map[string]*Log),
	}
	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

func Init(config *LogConfig) error {
	if instance.closed {
		return errors.New("logger already closed")
	}

	if config == nil {
		return errors.New("config is empty")
	}

	config, err := initConfig(config)
	if err != nil {
		return err
	}

	ma, err := NewMultiAppender(config.Outputs...)
	if err != nil {
		return err
	}

	var errs error
	newAppenders := make(map[string]Appender)
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
			_ = val.Close()
		}
		return errs
	}

	oldMa := instance.ma.Load()
	instance.ma.Store(ma)
	instance.level.Store(int32(config.Level))
	instance.closed = false

	for key, val := range oldMa.Appenders {
		if _, ok := ma.Appenders[key]; ok {
			continue
		}
		_ = val.Close()
	}

	return nil
}

func Close() error {
	instance.mu.Lock()
	if instance.closed {
		instance.mu.Unlock()
		return errors.New("logger already closed")
	}

	instance.closed = true
	close(instance.logTunnel)
	instance.mu.Unlock()

	instance.wg.Wait()
	return instance.ma.Load().Close()
}

func GetLog(packageName string) *Log {
	return logs.get(packageName)
}

func bootstrapWriter(app *appLogger) {
	defer app.wg.Done()

	for r := range app.logTunnel {
		if err := app.ma.Load().Write(r); err != nil {
			select {
			case app.errTunnel <- err:
			default:
				fmt.Fprintf(os.Stderr, "log4go: error tunnel overflow, message dropped: %q\n", err)
			}
		}
	}
	close(app.errTunnel)
}

func bootstrapError(app *appLogger) {
	defer app.wg.Done()

	for err := range app.errTunnel {
		fmt.Fprintf(os.Stderr, "log4go: Error - %q\n", err)
	}
}
