package log4go

import (
	"errors"
	"fmt"
)

type MultiAppender struct {
	Appenders map[string](Appender)
}

func NewMultiAppender(v ...*AppenderConfig) (*MultiAppender, error) {
	ma := &MultiAppender{}

	a := make(map[string](Appender))
	for _, val := range v {
		app, err := appenderFromConfig(val)
		if err != nil {
			return nil, err
		}
		a[app.Key()] = app
	}
	ma.Appenders = a
	return ma, nil
}

func (m *MultiAppender) Init() error {
	var errs error
	for _, a := range m.Appenders {
		err := a.Init()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (m *MultiAppender) Write(r *Record) error {
	var errs error
	for _, a := range m.Appenders {
		err := a.Write(r)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (m *MultiAppender) Close() error {
	var errs error
	for _, a := range m.Appenders {
		err := a.Close()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func appenderFromConfig(ac *AppenderConfig) (Appender, error) {
	var appender Appender
	switch ac.Type {
	case AppenderTypeConsole:
		appender = &ConsoleAppender{
			Target: ac.Target,
			Format: ac.Format,
		}
	case AppenderTypeFile:
	default:
		return nil, fmt.Errorf("wrong AppenderType: %q", ac.Type)
	}

	return appender, nil
}
