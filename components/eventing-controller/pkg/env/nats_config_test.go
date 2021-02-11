package env

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_GetNatsConfig(t *testing.T) {
	envs := map[string]string{
		"NATS_URL":          "NATS_URL",
		"EVENT_TYPE_PREFIX": "EVENT_TYPE_PREFIX",
	}

	g := NewGomegaWithT(t)
	defer func() {
		for k := range envs {
			err := os.Unsetenv(k)
			g.Expect(err).ShouldNot(HaveOccurred())
		}
	}()

	for k, v := range envs {
		err := os.Setenv(k, v)
		g.Expect(err).ShouldNot(HaveOccurred())
	}

	maxReconnects, reconnectWait := 1, time.Second
	config := GetNatsConfig(maxReconnects, reconnectWait)

	g.Expect(config.MaxReconnects).To(Equal(maxReconnects))
	g.Expect(config.ReconnectWait).To(Equal(reconnectWait))

	g.Expect(config.Url).To(Equal(envs["NATS_URL"]))
	g.Expect(config.EventTypePrefix).To(Equal(envs["EVENT_TYPE_PREFIX"]))
}
