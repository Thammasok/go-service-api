package pkg

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerLevelsText(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, InfoLevel, false)

	l.Debug("hidden", nil)
	l.Info("visible", map[string]any{"k": "v"})

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Fatalf("debug message should not be logged at info level")
	}
	if !strings.Contains(out, "visible") || !strings.Contains(out, "[INFO]") {
		t.Fatalf("expected info message in output, got: %q", out)
	}
}

func TestLoggerJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, DebugLevel, true)

	l.Error("oops", map[string]any{"code": 123})

	out := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(out), &obj); err != nil {
		t.Fatalf("output is not valid JSON: %v, out=%q", err, out)
	}
	if obj["level"] != "ERROR" || obj["msg"] != "oops" {
		t.Fatalf("unexpected json fields: %v", obj)
	}
	if fields, ok := obj["fields"].(map[string]any); !ok {
		t.Fatalf("expected fields object, got: %T", obj["fields"])
	} else if fields["code"].(float64) != 123 {
		t.Fatalf("expected code 123 in fields, got: %v", fields["code"])
	}
}

func TestWithFieldsMerged(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, InfoLevel, true)
	child := l.WithFields(map[string]any{"service": "api"})

	child.Info("started", map[string]any{"port": 8080})
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &obj); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	f, ok := obj["fields"].(map[string]any)
	if !ok {
		t.Fatalf("missing fields in entry: %v", obj)
	}
	if f["service"] != "api" {
		t.Fatalf("expected service field, got: %v", f["service"])
	}
	if f["port"].(float64) != 8080 {
		t.Fatalf("expected port 8080, got: %v", f["port"])
	}
}

func TestPackageHelpers(t *testing.T) {
	var buf bytes.Buffer
	std = NewLogger(&buf, InfoLevel, false)

	Debug("d", nil) // should be suppressed
	Info("i", nil)

	out := buf.String()
	if strings.Contains(out, "d") {
		t.Fatalf("debug should not be printed at info level")
	}
	if !strings.Contains(out, "i") {
		t.Fatalf("info should be printed, got: %q", out)
	}
}
