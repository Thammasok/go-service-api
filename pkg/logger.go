package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level controls log verbosity.
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

// Logger is a small, concurrency-safe leveled logger that can emit
// either human-readable text or JSON.
type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	level  Level
	json   bool
	fields map[string]any
}

// NewLogger constructs a new Logger writing to out with the provided level.
func NewLogger(out io.Writer, level Level, jsonFmt bool) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{
		out:    out,
		level:  level,
		json:   jsonFmt,
		fields: make(map[string]any),
	}
}

// NewDefault returns a basic logger to stdout at Info level.
func NewDefault() *Logger {
	return NewLogger(os.Stdout, InfoLevel, false)
}

func (l *Logger) clone() *Logger {
	nl := &Logger{
		out:   l.out,
		level: l.level,
		json:  l.json,
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
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// SetJSON toggles JSON output.
func (l *Logger) SetJSON(jsonFmt bool) {
	l.mu.Lock()
	l.json = jsonFmt
	l.mu.Unlock()
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}

	// merge logger fields and entry fields
	data := make(map[string]any, len(l.fields)+len(fields)+3)
	for k, v := range l.fields {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.json {
		entry := map[string]any{
			"time":  time.Now().Format(time.RFC3339),
			"level": level.String(),
			"msg":   msg,
		}
		if len(data) > 0 {
			entry["fields"] = data
		}
		enc, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(l.out, "{"+"\"time\":\"%s\",\"level\":\"%s\",\"msg\":\"%s\"}"+"\n", time.Now().Format(time.RFC3339), level.String(), msg)
			return
		}
		fmt.Fprintln(l.out, string(enc))
		return
	}

	// human readable
	out := fmt.Sprintf("%s [%s] %s", time.Now().Format(time.RFC3339), level.String(), msg)
	if len(data) > 0 {
		for k, v := range data {
			out += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	fmt.Fprintln(l.out, out)
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
