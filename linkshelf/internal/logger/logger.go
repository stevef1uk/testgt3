package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps a zap.Logger and provides convenience methods.
type Logger struct {
	zap *zap.Logger
}

// NewLogger creates a new Logger with the given log level.
// Supported levels: debug, info, warn, error, fatal.
func NewLogger(level string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	cfg.Encoding = "json"
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{zap: l}, nil
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.zap.Sugar().Info(msg)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	l.zap.Sugar().Warn(msg)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string) {
	l.zap.Sugar().Fatal(msg)
}

// WithFields adds structured fields to the logger and returns a new Logger.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	z := l.zap.With(zapFields...)
	return &Logger{zap: z}
}

// WithContext adds a map of fields to the logger context (same as WithFields).
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	return l.WithFields(fields)
}
