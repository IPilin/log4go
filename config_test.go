package log4go

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_log4go.yaml")

	yamlData := `level: "DEBUG"
hot_reload: false
outputs:
  - target: "stdout"
    format: "Text"
`

	err := os.WriteFile(configPath, []byte(yamlData), 0644)
	if err != nil {
		t.Fatalf("cannot write test config file: %q\n", configPath)
	}

	err = Init(&LogConfig{ConfigPath: configPath})
	if err != nil {
		t.Fatalf("error while InitConfig: %v\n", err)
	}
}
