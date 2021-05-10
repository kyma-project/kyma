package env

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_GetConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	envs := map[string]string{
		// required
		"CLIENT_ID":              "CLIENT_ID",
		"CLIENT_SECRET":          "CLIENT_SECRET",
		"TOKEN_ENDPOINT":         "TOKEN_ENDPOINT",
		"WEBHOOK_CLIENT_ID":      "WEBHOOK_CLIENT_ID",
		"WEBHOOK_CLIENT_SECRET":  "WEBHOOK_CLIENT_SECRET",
		"WEBHOOK_TOKEN_ENDPOINT": "WEBHOOK_TOKEN_ENDPOINT",
		"DOMAIN":                 "DOMAIN",
		"EVENT_TYPE_PREFIX":      "EVENT_TYPE_PREFIX",
		// optional
		"BEB_API_URL":                "BEB_API_URL",
		"BEB_NAMESPACE":              "/test",
		"WEBHOOK_ACTIVATION_TIMEOUT": "60s",
	}
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
	config := GetConfig()
	// Ensure required variables can be set
	g.Expect(config.ClientID).To(Equal(envs["CLIENT_ID"]))
	g.Expect(config.ClientSecret).To(Equal(envs["CLIENT_SECRET"]))
	g.Expect(config.TokenEndpoint).To(Equal(envs["TOKEN_ENDPOINT"]))
	g.Expect(config.WebhookClientID).To(Equal(envs["WEBHOOK_CLIENT_ID"]))
	g.Expect(config.WebhookClientSecret).To(Equal(envs["WEBHOOK_CLIENT_SECRET"]))
	g.Expect(config.WebhookTokenEndpoint).To(Equal(envs["WEBHOOK_TOKEN_ENDPOINT"]))
	g.Expect(config.Domain).To(Equal(envs["DOMAIN"]))
	g.Expect(config.EventTypePrefix).To(Equal(envs["EVENT_TYPE_PREFIX"]))
	g.Expect(config.BEBNamespace).To(Equal(envs["BEB_NAMESPACE"]))
	// Ensure optional variables can be set
	g.Expect(config.BebApiUrl).To(Equal(envs["BEB_API_URL"]))

	webhookActivationTimeout, err := time.ParseDuration(envs["WEBHOOK_ACTIVATION_TIMEOUT"])
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(config.WebhookActivationTimeout).To(Equal(webhookActivationTimeout))
}
