package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	// Global logger instance
	Logger zerolog.Logger

	// Log level mapping
	logLevels = map[string]zerolog.Level{
		"debug":    zerolog.DebugLevel,
		"info":     zerolog.InfoLevel,
		"warn":     zerolog.WarnLevel,
		"error":    zerolog.ErrorLevel,
		"fatal":    zerolog.FatalLevel,
		"panic":    zerolog.PanicLevel,
		"disabled": zerolog.Disabled,
		"trace":    zerolog.TraceLevel,
	}
)

// Config holds logger configuration
type Config struct {
	// Level is the minimum log level (debug, info, warn, error, fatal, panic)
	Level string `json:"level" env:"LOG_LEVEL" default:"info"`

	// Format is the output format (json, console)
	Format string `json:"format" env:"LOG_FORMAT" default:"json"`

	// Output destinations (stdout, stderr, file)
	Output string `json:"output" env:"LOG_OUTPUT" default:"stdout"`

	// FilePath for file output
	FilePath string `json:"file_path" env:"LOG_FILE_PATH" default:"logs/app.log"`

	// AddTimestamp adds timestamp to logs
	AddTimestamp bool `json:"add_timestamp" env:"LOG_TIMESTAMP" default:"true"`

	// AddCaller adds caller information to logs
	AddCaller bool `json:"add_caller" env:"LOG_CALLER" default:"true"`

	// PrettyPrint enables pretty printing for console format
	PrettyPrint bool `json:"pretty_print" env:"LOG_PRETTY" default:"false"`

	// ErrorStackTrace enables stack trace for errors
	ErrorStackTrace bool `json:"error_stack_trace" env:"LOG_STACK_TRACE" default:"true"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:           "info",
		Format:          "json",
		Output:          "stdout",
		FilePath:        "logs/app.log",
		AddTimestamp:    true,
		AddCaller:       true,
		PrettyPrint:     false,
		ErrorStackTrace: true,
	}
}

// Init initializes the global logger with configuration
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Configure zerolog global settings
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Set log level
	level, ok := logLevels[strings.ToLower(cfg.Level)]
	if !ok {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	case "file":
		if err := ensureLogDir(cfg.FilePath); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	case "stdout":
		fallthrough
	default:
		output = os.Stdout
	}

	// Configure format
	if cfg.Format == "console" {
		if cfg.PrettyPrint {
			output = zerolog.ConsoleWriter{
				Out:        output,
				TimeFormat: "15:04:05",
				NoColor:    false,
			}
		} else {
			output = zerolog.ConsoleWriter{
				Out:        output,
				TimeFormat: time.RFC3339,
				NoColor:    true,
			}
		}
	}

	// Create logger context
	logContext := zerolog.New(output)

	// Add timestamp if configured
	if cfg.AddTimestamp {
		logContext = logContext.With().Timestamp().Logger()
	}

	// Add caller if configured
	if cfg.AddCaller {
		logContext = logContext.With().Caller().Logger()
	}

	// Add hostname and PID for production environments
	if hostname, err := os.Hostname(); err == nil {
		logContext = logContext.With().Str("hostname", hostname).Logger()
	}
	logContext = logContext.With().Int("pid", os.Getpid()).Logger()

	// Set global logger
	Logger = logContext
	log.Logger = logContext

	return nil
}

// InitDevelopment initializes logger with development settings
func InitDevelopment() {
	Init(&Config{
		Level:           "debug",
		Format:          "console",
		Output:          "stdout",
		AddTimestamp:    true,
		AddCaller:       true,
		PrettyPrint:     true,
		ErrorStackTrace: true,
	})
}

// InitProduction initializes logger with production settings
func InitProduction() {
	Init(&Config{
		Level:           "info",
		Format:          "json",
		Output:          "stdout",
		AddTimestamp:    true,
		AddCaller:       true,
		PrettyPrint:     false,
		ErrorStackTrace: true,
	})
}

// ensureLogDir ensures the log directory exists
func ensureLogDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *zerolog.Logger {
	l := Logger.With().Logger()
	
	// Add request ID if present
	if reqID := GetRequestID(ctx); reqID != "" {
		l = l.With().Str("request_id", reqID).Logger()
	}
	
	// Add user ID if present
	if userID := GetUserID(ctx); userID != "" {
		l = l.With().Str("user_id", userID).Logger()
	}
	
	return &l
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) *zerolog.Logger {
	l := Logger.With().Fields(fields).Logger()
	return &l
}

// WithField returns a logger with an additional field
func WithField(key string, value interface{}) *zerolog.Logger {
	l := Logger.With().Interface(key, value).Logger()
	return &l
}

// WithError returns a logger with an error field
func WithError(err error) *zerolog.Logger {
	l := Logger.With().Err(err).Logger()
	return &l
}

// GetCaller returns the caller information
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Convenience functions that use the global logger

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	Logger.Debug().Msgf(format, v...)
}

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	Logger.Info().Msgf(format, v...)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	Logger.Warn().Msgf(format, v...)
}

// Error logs an error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	Logger.Error().Msgf(format, v...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	Logger.Fatal().Msgf(format, v...)
}

// Panic logs a panic message and panics
func Panic(msg string) {
	Logger.Panic().Msg(msg)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, v ...interface{}) {
	Logger.Panic().Msgf(format, v...)
}