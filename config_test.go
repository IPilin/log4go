package log4go

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadConfigFromFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePath := path.Join(tempDir, "log4go_temp")
	tempFile, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("error while creating test file: %v", tempFilePath)
	}
	defer tempFile.Close()

	expected := LogConfig{
		ConfigPath: tempFilePath,
		HotReload:  true,
		Level:      TraceLevel,
		Outputs: []*AppenderConfig{
			&AppenderConfig{
				Type: AppenderTypeFile,
				Path: tempDir,
			},
		},
	}

	tempConfig := fmt.Sprintf(`
hot_reload: true
level: trace
outputs:
  - type: file
    path: %q
`, tempDir)

	_, err = tempFile.WriteString(tempConfig)
	if err != nil {
		t.Fatalf("error while writing to test file: %v", tempFilePath)
	}

	got, err := loadFromFile(&expected)
	if err != nil {
		t.Fatalf("error loading config from file: %v", err)
	}

	if diff := cmp.Diff(expected, *got); diff != "" {
		t.Fatalf("config missmatch (-want +got):\n%s", diff)
	}
}

func TestLoadConfigFromFile_PreservesMissingFields(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePath := path.Join(tempDir, "log4go_temp")
	tempFile, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("error while creating test file: %v", tempFilePath)
	}
	defer tempFile.Close()

	_, err = tempFile.WriteString(`hot_reload: true`)
	if err != nil {
		t.Fatalf("error while writing to test file: %v", tempFilePath)
	}

	outputs := []*AppenderConfig{
		{
			Type:   AppenderTypeConsole,
			Target: TargetStderr,
			Format: FormatText,
		},
	}
	config := &LogConfig{
		ConfigPath: tempFilePath,
		HotReload:  false,
		Level:      ErrorLevel,
		Outputs:    outputs,
	}

	got, err := loadFromFile(config)
	if err != nil {
		t.Fatalf("error loading config from file: %v", err)
	}

	want := LogConfig{
		ConfigPath: tempFilePath,
		HotReload:  true,
		Level:      ErrorLevel,
		Outputs:    outputs,
	}
	if diff := cmp.Diff(want, *got); diff != "" {
		t.Fatalf("config missmatch (-want +got):\n%s", diff)
	}
}

func TestLoadConfigFromFile_Error(t *testing.T) {
	tests := []struct {
		name   string
		action func(t *testing.T)
	}{
		{
			name: "nil config",
			action: func(t *testing.T) {
				got, err := loadFromFile(nil)
				if err == nil {
					t.Fatal("expected error got nil")
				}
				if got != nil {
					t.Fatalf("expected nil got: %v", got)
				}
			},
		},
		{
			name: "wrong file path",
			action: func(t *testing.T) {
				config := &LogConfig{
					ConfigPath: "",
				}

				got, err := loadFromFile(config)
				if err == nil {
					t.Fatal("expected error got nil")
				}
				if got != nil {
					t.Fatalf("expected nil got: %v", got)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.action)
	}
}
