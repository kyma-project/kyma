package logger

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	t.Run("should proxy requests", func(t *testing.T) {
		logger := New(TEXT, DEBUG)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		testZapLogger(logger)
		time.Sleep(time.Second * 5)

		logs := buffer.String()
		logsSlice := strings.Split(logs, "\n")
		require.NotEqual(t, 0, len(logsSlice), "there are no logs")
		t.Log(logs)
	})
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func testZapLogger(log *zap.SugaredLogger) {
	log.With("context", "a", "a").Infof("just normal log with msg: %s", "Hello From Zap")
	log.Errorf("Error msg: %s", "some error occurred")
}
