package log4go

import (
	"errors"
	"os"
	"path/filepath"
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

func TestAppender_Init_StandardStreams(t *testing.T) {
	appStdout := &Appender{Target: Stdout}
	if err := appStdout.Init(); err != nil {
		t.Fatalf("unexpected error initializing stdout appender: %v", err)
	}
	if appStdout.writer != os.Stdout {
		t.Errorf("expected writer to be os.Stdout")
	}

	appStderr := &Appender{Target: Stderr}
	if err := appStderr.Init(); err != nil {
		t.Fatalf("unexpected error initializing stderr appender: %v", err)
	}
	if appStderr.writer != os.Stderr {
		t.Errorf("expected writer to be os.Stderr")
	}
}

func TestAppender_File_Lifecycle(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	appender := &Appender{
		Target: File,
		Path:   tmpFile,
	}

	// Test Init
	if err := appender.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if appender.file == nil {
		t.Fatal("expected appender.file to be non-nil")
	}
	if appender.writer == nil {
		t.Fatal("expected appender.writer to be non-nil")
	}

	// Test Write
	msg := []byte("hello log4go")
	n, err := appender.Write(msg)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(msg) {
		t.Errorf("expected to write %d bytes, wrote %d", len(msg), n)
	}

	// Test Close
	if err := appender.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}
	if string(content) != string(msg) {
		t.Errorf("expected file content %q, got %q", string(msg), string(content))
	}
}

func TestAppender_Init_File_Error(t *testing.T) {
	appender := &Appender{
		Target: File,
		Path:   filepath.Join("invalid", "directory", "path", "that", "does", "not", "exist.log"),
	}

	err := appender.Init()
	if err == nil {
		t.Error("expected error initializing file appender with invalid path, got nil")
	}
}

func TestAppender_Write_NoWriterError(t *testing.T) {
	appender := &Appender{
		Target: File,
		Path:   "dummy.log",
		file:   &os.File{}, // force file != nil
		writer: nil,        // force writer == nil
	}

	_, err := appender.Write([]byte("test"))
	if err == nil {
		t.Error("expected error when writing without writer, got nil")
	}
}

func TestMultiAppender_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "log1.log")
	file2 := filepath.Join(tmpDir, "log2.log")

	appender1 := &Appender{Target: File, Path: file1}
	appender2 := &Appender{Target: File, Path: file2}

	ma := NewMultiAppender(appender1, appender2)

	// Test Write
	msg := []byte("multi-appender test")
	n, err := ma.Write(msg)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(msg) {
		t.Errorf("expected to write %d bytes, wrote %d", len(msg), n)
	}

	// Test Close
	if err := ma.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify contents of both files
	for _, f := range []string{file1, file2} {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("could not read log file %q: %v", f, err)
		}
		if string(content) != string(msg) {
			t.Errorf("expected file content %q in %q, got %q", string(msg), f, string(content))
		}
	}
}

func TestMultiAppender_Errors(t *testing.T) {
	invalidAppender := &Appender{
		Target: File,
		Path:   filepath.Join("invalid", "path", "1.log"),
	}
	invalidAppender2 := &Appender{
		Target: File,
		Path:   filepath.Join("invalid", "path", "2.log"),
	}

	// Init error
	maInit := &MultiAppender{Appenders: []*Appender{invalidAppender, invalidAppender2}}
	if err := maInit.Init(); err == nil {
		t.Error("expected error from MultiAppender.Init, got nil")
	}

	// Write error
	badWriteAppender := &Appender{Target: File, Path: "dummy.log", file: &os.File{}, writer: nil}
	maWrite := &MultiAppender{Appenders: []*Appender{badWriteAppender}}
	if _, err := maWrite.Write([]byte("test")); err == nil {
		t.Error("expected error from MultiAppender.Write, got nil")
	}

	// Close error (simulate by trying to close a pre-closed temp file)
	tempFile, _ := os.CreateTemp("", "test-close")
	tempFile.Close() // close immediately

	badCloseAppender := &Appender{Target: File, file: tempFile}
	maClose := &MultiAppender{Appenders: []*Appender{badCloseAppender}}
	if err := maClose.Close(); err == nil {
		t.Error("expected error from MultiAppender.Close, got nil")
	}
	os.Remove(tempFile.Name())
}
