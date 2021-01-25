package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type logEntry struct {
	Context   map[string]string `json:"context"`
	Msg       string            `json:"message"`
	TraceID   string            `json:"traceid"`
	SpanID    string            `json:"spanid"`
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
}

func TestLogger(t *testing.T) {
	t.Run("should log anything", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(zap.DebugLevel)
		log := logger.New(logger.JSON, logger.DEBUG, core).WithContext()

		// when
		log.Desugar().WithOptions(zap.AddCaller())
		log.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})

	t.Run("should log in the right json format", func(t *testing.T) {
		// GIVEN
		array := make([]byte, 0)
		buffer := bytes.NewBuffer(array)

		syncBuffer := zapcore.AddSync(buffer)
		ws := zapcore.Lock(syncBuffer)
		logFilter := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return true
		})

		core := zapcore.NewCore(logger.JSON.ToZapEncoder(), ws, logFilter)
		log := logger.New(logger.JSON, logger.DEBUG, core)

		ctx := fixContext(map[string]string{"traceid": "trace", "spanid": "span"})
		// WHEN
		log.WithTracing(ctx).With("key", "value").Info("example message")

		// THEN
		require.NotEqual(t, 0, buffer.Len())
		var entry = logEntry{}
		strictEncoder := json.NewDecoder(strings.NewReader(buffer.String()))
		strictEncoder.DisallowUnknownFields()
		err := strictEncoder.Decode(&entry)
		require.NoError(t, err)

		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "example message", entry.Msg)
		assert.Equal(t, "trace", entry.TraceID)
		assert.Equal(t, "span", entry.SpanID)

		assert.NotEmpty(t, entry.Timestamp)
		_, err = time.Parse(time.RFC3339, entry.Timestamp)
		assert.NoError(t, err)
	})

	t.Run("should log in total separation", func(t *testing.T) {
		array := make([]byte, 0)
		buffer := bytes.NewBuffer(array)

		syncBuffer := zapcore.AddSync(buffer)
		ws := zapcore.Lock(syncBuffer)
		logFilter := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return true
		})
		core := zapcore.NewCore(logger.JSON.ToZapEncoder(), ws, logFilter)
		log := logger.New(logger.JSON, logger.DEBUG, core)
		ctx := fixContext(map[string]string{"traceid": "trace", "spanid": "span"})

		// WHEN
		log.WithTracing(ctx).With("key", "first").Info("first message")
		log.WithContext().With("key", "second").Error("second message")

		// THEN
		require.NotEqual(t, 0, buffer.Len())

		logs := strings.Split(string(buffer.Bytes()), "\n")

		require.Len(t, logs, 3) // 3rd line is new empty line

		var infoEntry = logEntry{}
		strictEncoder := json.NewDecoder(strings.NewReader(logs[0]))
		strictEncoder.DisallowUnknownFields()
		err := strictEncoder.Decode(&infoEntry)
		require.NoError(t, err)

		assert.Equal(t, "INFO", infoEntry.Level)
		assert.Equal(t, "first message", infoEntry.Msg)
		assert.EqualValues(t, map[string]string{"key": "first"}, infoEntry.Context, 0.0)
		assert.Equal(t, "span", infoEntry.SpanID)
		assert.Equal(t, "trace", infoEntry.TraceID)

		assert.NotEmpty(t, infoEntry.Timestamp)
		_, err = time.Parse(time.RFC3339, infoEntry.Timestamp)
		assert.NoError(t, err)

		strictEncoder = json.NewDecoder(strings.NewReader(logs[1]))
		strictEncoder.DisallowUnknownFields()

		var errorEntry = logEntry{}
		err = strictEncoder.Decode(&errorEntry)
		assert.Equal(t, "ERROR", errorEntry.Level)
		assert.Equal(t, "second message", errorEntry.Msg)
		assert.EqualValues(t, map[string]string{"key": "second"}, errorEntry.Context, 0.0)
		assert.Empty(t, errorEntry.SpanID)
		assert.Empty(t, errorEntry.TraceID)

		assert.NotEmpty(t, errorEntry.Timestamp)
		_, err = time.Parse(time.RFC3339, errorEntry.Timestamp)
		assert.NoError(t, err)
	})

	t.Run("with context should create new logger", func(t *testing.T) {
		//GIVEN
		log := logger.New(logger.TEXT, logger.INFO)

		//WHEN
		firstLogger := log.WithContext()
		secondLogger := log.WithContext()

		//THEN
		assert.NotSame(t, firstLogger, secondLogger)
	})

	t.Run("with tracing should create new logger", func(t *testing.T) {
		//GIVEN
		log := logger.New(logger.TEXT, logger.INFO)
		ctx := fixContext(map[string]string{"traceid": "trace", "spanid": "span"})

		//WHEN
		firstLogger := log.WithTracing(ctx)
		secondLogger := log.WithTracing(ctx)

		//THEN
		assert.NotSame(t, firstLogger, secondLogger)
	})
}

func fixContext(values map[string]string) context.Context {
	ctx := context.TODO()
	for k, v := range values {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}
