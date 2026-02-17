package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerLevelsText(t *testing.T) {
	t.Run("info level output", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, InfoLevel, false)

		l.Debug("hidden", nil)
		l.Info("visible", map[string]any{"k": "v"})

		out := buf.String()
		assert.NotContains(t, out, "hidden", "debug message should not be logged at info level")
		assert.Contains(t, out, "visible", "info message should be in output")
		assert.Contains(t, out, "msg=visible", "output should contain msg field")
	})
}

func TestLoggerJSONOutput(t *testing.T) {
	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, DebugLevel, true)

		l.Error("oops", map[string]any{"code": 123})

		var obj map[string]any
		err := json.Unmarshal([]byte(buf.String()), &obj)
		require.NoError(t, err, "output should be valid JSON")

		// Logrus uses lowercase "error" for error level
		assert.Equal(t, "error", obj["level"], "log level should be 'error'")
		assert.Equal(t, "oops", obj["msg"], "log message should be 'oops'")

		// Logrus merges fields directly into the entry
		assert.Equal(t, float64(123), obj["code"], "code field should be 123")
	})
}

func TestWithFieldsMerged(t *testing.T) {
	t.Run("fields merged correctly", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, InfoLevel, true)
		child := l.WithFields(map[string]any{"service": "api"})

		child.Info("started", map[string]any{"port": 8080})

		var obj map[string]any
		err := json.Unmarshal([]byte(buf.String()), &obj)
		require.NoError(t, err, "output should be valid JSON")

		// Logrus merges fields directly into the entry
		assert.Equal(t, "api", obj["service"], "service field should be 'api'")
		assert.Equal(t, float64(8080), obj["port"], "port field should be 8080")
		assert.Equal(t, "started", obj["msg"], "message should be 'started'")
	})
}

func TestPackageHelpers(t *testing.T) {
	t.Run("package level helpers", func(t *testing.T) {
		var buf bytes.Buffer
		std = NewLogger(&buf, InfoLevel, false)

		Debug("d", nil) // should be suppressed
		Info("i", nil)

		out := buf.String()
		assert.NotContains(t, out, "d", "debug should not be printed at info level")
		assert.Contains(t, out, "i", "info should be printed")
	})
}
