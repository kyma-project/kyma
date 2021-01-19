package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		core, observedLogs := observer.New(DEBUG.toZapLevel())
		logger := New(JSON, DEBUG, core).WithContext()

		// when
		logger.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})

	t.Run("should log in the right format", func(t *testing.T) {
		// given

		array := make([]byte, 0)
		buffer := bytes.NewBuffer(array)

		syncBuffer := zapcore.AddSync(buffer)
		ws := zapcore.Lock(syncBuffer)
		logFilter := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return true
		})

		core := zapcore.NewCore(JSON.ToZapEncoder(), ws, logFilter)
		logger := New(JSON, DEBUG, core)

		ctx := fixContext(map[string]string{"traceid": "trace", "spanid": "span"})
		// when
		logger.WithTracing(ctx).With("key", "value").Info("example message")

		// then
		var entry = logEntry{}
		require.NotEqual(t, 0, buffer.Len())
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)
		//MySuper
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "example message", entry.Msg)
		assert.Equal(t, "trace", entry.TraceID)
		assert.Equal(t, "span", entry.SpanID)

		assert.NotEmpty(t, entry.Timestamp)
		_, err = time.Parse(time.RFC3339, entry.Timestamp)
		assert.NoError(t, err)

	})
}

func fixContext(values map[string]string) context.Context {
	ctx := context.TODO()
	for k, v := range values {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}

//func containsZapField(field zapcore.Field, fields []zapcore.Field) bool {
//	for _, val := range fields {
//		if field.Equals(val) {
//			return true
//		}
//	}
//	return false
//}

//func TestParallelRun(t *testing.T) {
//	goRuntimesNumber := 15
//	logger := New(TEXT, INFO)
//
//	wg := sync.WaitGroup{}
//
//	for i := 0; i < goRuntimesNumber; i++ {
//
//		wg.Add(1)
//		i := i
//		go func() {
//			defer wg.Done()
//			fmt.Println("aaaa")
//			logger.WithContext(map[string]string{
//				"id": strconv.Itoa(i),
//			}).Infof("Hello From: %d", i)
//		}()
//	}
//
//	wg.Wait()
//
//}
