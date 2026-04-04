package log4go

import (
	"errors"
	"testing"
)

func TestLogLevel_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{"Trace uppercase", "TRACE", TraceLevel},
		{"Debug uppercase", "DEBUG", DebugLevel},
		{"Info uppercase", "INFO", InfoLevel},
		{"Warn uppercase", "WARN", WarnLevel},
		{"Error uppercase", "ERROR", ErrorLevel},
		{"Lowercase case-insensitive", "debug", DebugLevel},
		{"Mixed case", "InFo", InfoLevel},
		{"Unknown fallback to INFO", "UNKNOWN", InfoLevel},
		{"Empty string fallback to INFO", "", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var level LogLevel
			err := level.UnmarshalYAML(func(v any) error {
				*v.(*string) = tt.input
				return nil
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if level != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestLogLevel_UnmarshalYAML_Error(t *testing.T) {
	var level LogLevel
	expectedErr := errors.New("mock unmarshal error")

	err := level.UnmarshalYAML(func(v any) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
