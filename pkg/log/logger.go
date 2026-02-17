package log

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger with service metadata
type Logger struct {
	*zap.Logger
}

var globalLogger *Logger

// LogLevel represents log levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// GetLogLevelFromString converts a string log level to LogLevel
func GetLogLevelFromString(level string) LogLevel {
	level = strings.ToUpper(strings.TrimSpace(level))
	switch level {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// getZapLevel converts our LogLevel to zap's Level
func getZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := os.Getenv("ENV")
	if env != "" {
		return env
	}
	return "local"
}

// NewLogger creates a new logger instance with service metadata
func NewLogger(serviceName string, level LogLevel) (*Logger, error) {
	zapLevel := getZapLevel(level)

	// Configure encoder for JSON output
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.TimeKey = "ts"

	// Get service metadata
	hostname, _ := os.Hostname()
	version := os.Getenv("tag")
	if version == "" {
		version = "unknown"
	}

	// Create zap core with JSON encoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		zap.NewAtomicLevelAt(zapLevel),
	)

	// Create logger with service metadata as initial fields
	logger := zap.New(core).With(
		zap.String("service.name", serviceName),
		zap.String("service.id", hostname),
		zap.String("service.version", version),
		zap.String("service.env", getEnvironment()),
	)

	globalLogger = &Logger{Logger: logger}

	return globalLogger, nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// WithContext returns a zap logger with trace IDs extracted from context
// This is the key function that extracts trace_id, span_id, and request.id
// from the OpenTelemetry context and adds them as zap fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	fields := []zap.Field{}

	// STEP 2.1: Extract trace and span IDs from OpenTelemetry span context
	// OpenTelemetry stores span context in the context.Context
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		traceID := spanCtx.TraceID().String()
		spanID := spanCtx.SpanID().String()

		if traceID != "" {
			// Add both standard and Datadog format trace IDs
			fields = append(fields, zap.String("trace_id", traceID))
			fields = append(fields, zap.String("dd.trace_id", traceID))
		}
		if spanID != "" {
			// Add both standard and Datadog format span IDs
			fields = append(fields, zap.String("span_id", spanID))
			fields = append(fields, zap.String("dd.span_id", spanID))
		}
	}

	// STEP 2.2: Extract request ID from context
	// Request ID is set by Echo middleware and stored in context
	if reqID, ok := ctx.Value("requestID").(string); ok && reqID != "" {
		fields = append(fields, zap.String("request.id", reqID))
	} else if reqID, ok := ctx.Value("X-Request-ID").(string); ok && reqID != "" {
		fields = append(fields, zap.String("request.id", reqID))
	}

	// STEP 2.3: If we found any trace/request IDs, add them to logger
	if len(fields) > 0 {
		return l.Logger.With(fields...)
	}

	// If no trace IDs found, return logger as-is
	return l.Logger
}

// Convenience methods that automatically extract trace IDs from context

// Info logs an info message with trace IDs from context
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

// Warn logs a warn message with trace IDs from context
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

// Error logs an error message with trace IDs from context
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Error(msg, fields...)
}

// Debug logs a debug message with trace IDs from context
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// Fatal logs a fatal message with trace IDs from context
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Fatal(msg, fields...)
}

// Infof logs with format string (similar to log.Infof but with trace IDs)
// Convenience method for migration from gommon/log
func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Sugar().Infof(format, args...)
}

// Errorf logs with format string (similar to log.Errorf but with trace IDs)
func (l *Logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Sugar().Errorf(format, args...)
}

// Warnf logs with format string (similar to log.Warnf but with trace IDs)
func (l *Logger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Sugar().Warnf(format, args...)
}

// Debugf logs with format string (similar to log.Debugf but with trace IDs)
func (l *Logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Sugar().Debugf(format, args...)
}
