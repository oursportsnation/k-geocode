package logger

import (
	"fmt"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger with the specified level and format
func New(level string, format string) (*zap.Logger, error) {
	var config zap.Config
	
	// 기본 설정에 따라 config 선택
	if format == "console" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	
	// 로그 레벨 설정
	logLevel, err := parseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)
	
	// 출력 형식 설정
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	
	// Logger 생성
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	
	return logger, nil
}

// parseLevel converts string log level to zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// NewNop returns a no-op logger (for testing)
func NewNop() *zap.Logger {
	return zap.NewNop()
}