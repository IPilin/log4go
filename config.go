package log4go

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func initConfig(config *LogConfig) (*LogConfig, error) {
	var err error
	if config.ConfigPath != "" {
		config, err = loadFromFile(config)
		if err != nil {
			return nil, err
		}
	}

	if len(config.Outputs) == 0 {
		config.Outputs = append(config.Outputs, &Appender{
			Target: Stdout,
			Format: Text,
		})
		fmt.Printf("log4go: outputs are empty, using default output: %q\n", Stdout)
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
