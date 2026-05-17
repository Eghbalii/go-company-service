package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger configured for the given level and format.
// format: "json" (default) or "console" (human-readable for local dev).
func New(level, format string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.DisableStacktrace = true

	if format == "console" {
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return log, nil
}
