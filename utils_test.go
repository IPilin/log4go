package log4go

import (
	"bytes"
	"testing"
	"time"
)

func TestAppendFormatTime(t *testing.T) {
	var buf bytes.Buffer
	ts := time.Date(2026, 7, 3, 14, 5, 9, 0, time.UTC)

	appendFormatTime(&buf, ts)

	got := buf.String()
	want := "2026-07-03 14:05:09"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
