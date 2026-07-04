package log4go

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want LogLevel
	}{
		{"trace", `TRACE`, TraceLevel},
		{"debug lowercase", `debug`, DebugLevel},
		{"unknown level", `PRism`, InfoLevel},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var level LogLevel
			buf := []byte(test.in)
			err := yaml.Unmarshal(buf, &level)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if level != test.want {
				t.Fatalf("got %v, want %v", level, test.want)
			}
		})
	}
}
