package logger

import (
	"io"
	"maps"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Level is a log level type.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// toLogrusLevel converts our Level to logrus.Level.
func toLogrusLevel(l Level) logrus.Level {
	switch l {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// newTextFormatter returns a logrus TextFormatter with a clean, readable layout.
// colors=true enables ANSI colour codes (intended for interactive terminals).
func newTextFormatter(colors bool) logrus.Formatter {
	return &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     colors,
		DisableColors:   !colors,
		PadLevelText:    true,
	}
}

// Logger wraps logrus.Logger for consistent API.
type Logger struct {
	logrus *logrus.Logger
	fields map[string]any
}

// NewLogger constructs a new Logger using logrus backend.
func NewLogger(out io.Writer, level Level, jsonFmt bool) *Logger {
	if out == nil {
		out = os.Stdout
	}

	l := logrus.New()
	l.SetOutput(out)
	l.SetLevel(toLogrusLevel(level))

	if jsonFmt {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	} else {
		l.SetFormatter(newTextFormatter(false))
	}

	return &Logger{
		logrus: l,
		fields: make(map[string]any),
	}
}

// NewDefault returns a basic logger to stdout at Info level.
func NewDefault() *Logger {
	return NewLogger(os.Stdout, InfoLevel, false)
}

func (l *Logger) clone() *Logger {
	nl := &Logger{
		logrus: l.logrus,
	}
	nl.fields = make(map[string]any, len(l.fields))
	for k, v := range l.fields {
		nl.fields[k] = v
	}
	return nl
}

// WithFields returns a child logger that includes the provided fields
// on every log entry.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	nl := l.clone()
	for k, v := range fields {
		nl.fields[k] = v
	}
	return nl
}

// SetLevel updates the logger level.
func (l *Logger) SetLevel(level Level) {
	l.logrus.SetLevel(toLogrusLevel(level))
}

// SetJSON toggles JSON output.
func (l *Logger) SetJSON(jsonFmt bool) {
	if jsonFmt {
		l.logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	} else {
		l.logrus.SetFormatter(newTextFormatter(false))
	}
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	data := make(map[string]any, len(l.fields)+len(fields))
	maps.Copy(data, l.fields)
	for k, v := range fields {
		data[k] = v
	}

	entry := l.logrus.WithFields(data)

	switch level {
	case DebugLevel:
		entry.Debug(msg)
	case InfoLevel:
		entry.Info(msg)
	case WarnLevel:
		entry.Warn(msg)
	case ErrorLevel:
		entry.Error(msg)
	}
}

// Debug logs a message at Debug level.
func (l *Logger) Debug(msg string, fields map[string]any) { l.log(DebugLevel, msg, fields) }

// Info logs a message at Info level.
func (l *Logger) Info(msg string, fields map[string]any) { l.log(InfoLevel, msg, fields) }

// Warn logs a message at Warn level.
func (l *Logger) Warn(msg string, fields map[string]any) { l.log(WarnLevel, msg, fields) }

// Error logs a message at Error level.
func (l *Logger) Error(msg string, fields map[string]any) { l.log(ErrorLevel, msg, fields) }

// Package-level default logger.
var std = NewDefault()

// Std returns the package-level logger (for advanced callers).
func Std() *Logger { return std }

// Convenience helpers using the package default logger.
func Debug(msg string, fields map[string]any) { std.Debug(msg, fields) }
func Info(msg string, fields map[string]any)  { std.Info(msg, fields) }
func Warn(msg string, fields map[string]any)  { std.Warn(msg, fields) }
func Error(msg string, fields map[string]any) { std.Error(msg, fields) }

// SetLevel updates the default logger level.
func SetLevel(level Level) { std.SetLevel(level) }

// SetJSON toggles JSON output on the default logger.
func SetJSON(jsonFmt bool) { std.SetJSON(jsonFmt) }

// LevelFromEnv returns the log level appropriate for the given environment name.
//
//	development, local → DebugLevel  (all logs)
//	staging, test      → InfoLevel
//	production         → ErrorLevel
//	(anything else)    → InfoLevel
func LevelFromEnv(env string) Level {
	switch strings.ToLower(env) {
	case "development", "local":
		return DebugLevel
	case "staging", "test":
		return InfoLevel
	case "production":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// InitFromEnv configures the default logger for the given application environment.
// It sets the log level automatically and enables coloured output for
// development/local environments.
func InitFromEnv(env string) {
	std.SetLevel(LevelFromEnv(env))

	isDev := strings.ToLower(env) == "development" || strings.ToLower(env) == "local"
	std.logrus.SetFormatter(newTextFormatter(isDev))
}
