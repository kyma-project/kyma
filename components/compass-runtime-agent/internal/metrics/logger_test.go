package metrics

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
	metricsFake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	loggingInterval = time.Millisecond
	loggingWaitTime = time.Millisecond * 10
)

func Test_Log(t *testing.T) {
	t.Run("should log metrics", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset()
		metricsClientset := metricsFake.NewSimpleClientset() // TODO: WTF. It does not work ;-;

		logger := NewMetricsLogger(resourcesClientset, metricsClientset, loggingInterval)

		quitChannel := make(chan bool, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go logger.Log(quitChannel)

		time.Sleep(loggingWaitTime)
		quitChannel <- true
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		t.Log(logs) // TODO: Delete it!
		assert.Equal(t, true, strings.Contains(logs, "Cluster metrics logged successfully."), "did not log metrics")
		assert.Equal(t, true, strings.Contains(logs, "Logging stopped."), "did not finish gracefully")
		assert.Equal(t, true, strings.Contains(logs, "\"shouldBeFetched\":true"), "shouldBeFetched flag is not true")
		assert.Equal(t, false, strings.Contains(logs, "error"), "logged an error")
	})
}
