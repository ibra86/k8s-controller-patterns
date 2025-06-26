package cmd

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  zerolog.Level
	}{
		{"info", zerolog.InfoLevel},
		{"debug", zerolog.DebugLevel},
		{"trace", zerolog.TraceLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"unknown", zerolog.InfoLevel},
	}
	for _, tt := range tests {
		got := parseLogLevel(tt.input)
		if got != tt.want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
