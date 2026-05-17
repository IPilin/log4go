package log4go

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
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

type AppenderConfig struct {
	Type   AppenderType `yaml:"type"`
	Target Target       `yaml:"target"`
	Path   string       `yaml:"path"`
	Format OutputFormat `yaml:"format"`
}

type LogConfig struct {
	ConfigPath string
	HotReload  bool              `yaml:"hot_reload"`
	Level      LogLevel          `yaml:"level"`
	Outputs    []*AppenderConfig `yaml:"outputs"`
}

func initConfig(config *LogConfig) (*LogConfig, error) {
	var err error
	if config.ConfigPath != "" {
		config, err = loadFromFile(config)
		if err != nil {
			return nil, err
		}
	}

	if len(config.Outputs) == 0 {
		config.Outputs = append(config.Outputs, &AppenderConfig{
			Type:   AppenderTypeConsole,
			Target: TargetStdout,
			Format: FormatText,
		})
		fmt.Printf("log4go: outputs are empty, using default output: %q\n", TargetStdout)
	}

	return config, err
}

func loadFromFile(config *LogConfig) (*LogConfig, error) {
	file, err := os.Open(config.ConfigPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	newConfig := &LogConfig{
		ConfigPath: config.ConfigPath,
		HotReload:  config.HotReload,
	}

	err = yaml.Unmarshal(data, newConfig)
	if err != nil {
		return nil, err
	}

	return newConfig, nil
}
