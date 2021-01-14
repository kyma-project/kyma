package logger

import (
	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest/observer"
	"testing"
)

func TestLogger(t *testing.T) {
	t.Run("should log anything", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(DEBUG.toZapLevel())
		logger := New(JSON, DEBUG, core)

		// when
		logger.Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})

	t.Run("should log in the right format", func(t *testing.T) {
		// given
		core, observedLogs := observer.New(DEBUG.toZapLevel())
		logger := New(JSON, DEBUG, core)

		// when
		logger.WithTracing(map[string]string{"traceid": "traceid123", "spanid": "spanid123"}).
			WithContext(map[string]string{"key": "val"}).
			Debug("something")

		// then
		require.NotEqual(t, 0, observedLogs.Len())
		for _, log := range observedLogs.All() {
			//TODO: Consider some better approach
			assert.Equal(t, DEBUG.toZapLevel(), log.Level)
			assert.Equal(t, "something", log.Message)
			//fields := []zapcore.Field{{
			//	"traceid",
			//	0,
			//	0,
			//	"traceid123",
			//	nil,
			//},{
			//	"spanid",
			//	0,
			//	0,
			//	"spanid123",
			//	nil,
			//},{
			//	"traceid",
			//	0,
			//	0,
			//	"traceid123",
			//	nil,
			//},}
			//for _, val := range fields {
			//	assert.Equal(t, true, containsZapField(val, log.Context))
			//}
		}
	})
}

//func containsZapField(field zapcore.Field, fields []zapcore.Field) bool {
//	for _, val := range fields {
//		if field.Equals(val) {
//			return true
//		}
//	}
//	return false
//}
