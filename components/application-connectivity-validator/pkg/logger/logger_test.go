package logger

import (
	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest/observer"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	t.Run("should log anything", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(DEBUG.toZapLevel())
		logger := newWithCustomCores(JSON, DEBUG, core)

		// when
		logger.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})

	t.Run("should log in the right format", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(DEBUG.toZapLevel())
		logger := newWithCustomCores(JSON, DEBUG, core)

		// when
		//TODO: Here you can find out what actually we should do to the logger
		//		to perform such super unified log as we want to
		logger.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		for _, log := range observedLogs.All() {
			//TODO: It does not want to work this way. Try something else
			m := log.ContextMap()
			require.Contains(t, m, "timestamp")
			require.Contains(t, m, "level")
			require.Contains(t, m, "message")
			require.Contains(t, m, "context")
			require.Contains(t, m, "traceid")
			require.Contains(t, m, "spanid")

			assert.Equal(t, DEBUG.toZapLevel(), log.Level)

			if val, ok := m["timestamp"]; ok {
				_, err := time.Parse(time.RFC3339, val.(string))
				assert.Equal(t, nil, err)
			}
			if val, ok := m["level"]; ok {
				assert.Equal(t, string(DEBUG), val.(string))
			}
			if val, ok := m["message"]; ok {
				assert.Equal(t, "something", val.(string))
			}
			if val, ok := m["context"]; ok {
				assert.Equal(t, map[string]string{"key": "val"}, val.(map[string]string))
			}
			if val, ok := m["traceid"]; ok {
				assert.Equal(t, "traceid123", val.(string))
			}
			if val, ok := m["spanid"]; ok {
				assert.Equal(t, "spanid123", val.(string))
			}
		}
	})
}
