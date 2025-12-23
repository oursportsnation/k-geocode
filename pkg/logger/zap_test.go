package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		wantErr bool
	}{
		{"json format with info level", "info", "json", false},
		{"json format with debug level", "debug", "json", false},
		{"json format with warn level", "warn", "json", false},
		{"json format with error level", "error", "json", false},
		{"console format with info level", "info", "console", false},
		{"console format with debug level", "debug", "console", false},
		{"invalid level", "invalid", "json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, tt.format)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, logger)
				assert.Contains(t, err.Error(), "invalid log level")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, logger)
				logger.Sync()
			}
		})
	}
}

func TestNewNop(t *testing.T) {
	logger := NewNop()

	require.NotNil(t, logger)
	// Should not panic when logging
	logger.Info("test message")
	logger.Debug("debug message")
	logger.Warn("warn message")
	logger.Error("error message")
}
