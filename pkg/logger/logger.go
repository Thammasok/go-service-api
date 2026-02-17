package logger

import (
	"io"
	"os"

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

// Logger wraps logrus.Logger for consistent API.
type Logger struct {
	logrus *logrus.Logger
	fields map[string]any
}

// NewLogger constructs a new Logger using logrus backend with JSON formatting.
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
		l.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			FullTimestamp:   true,
		})
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
		l.logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			FullTimestamp:   true,
		})
	}
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	// merge logger fields and entry fields
	data := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		data[k] = v
	}
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

// Package-level default logger and helpers
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
