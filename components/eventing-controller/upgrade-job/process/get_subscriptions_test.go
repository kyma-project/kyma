package process

import (
	"log"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/onsi/gomega"
)

// TestGetSubscriptions tests the TestGetSubscriptions_DO step
func TestGetSubscriptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	e2eSetup := newE2ESetup()
	cfg := e2eSetup.config

	// Create logger instance
	ctrLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger in testing: %s", err)
	}

	// Create process
	p := &Process{
		Logger:         ctrLogger.Logger,
		TimeoutPeriod:  60 * time.Second,
		ReleaseName:    cfg.ReleaseName,
		KymaNamespace:  cfg.KymaNamespace,
		ControllerName: cfg.EventingControllerName,
		PublisherName:  cfg.EventingPublisherName,
		State:          State{},
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Get Kyma subscriptions", func(t *testing.T) {
		// First check if there are 0 subscriptions process state
		g.Expect(p.State.Subscriptions).Should(gomega.BeNil())

		// Now get the subscriptions
		p.Steps = []Step{
			NewGetSubscriptions(p),
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if there are 4 subscriptions process state
		g.Expect(p.State.Subscriptions).Should(gomega.Not(gomega.BeNil()))
		g.Expect(p.State.Subscriptions.Items).Should(gomega.HaveLen(4))
	})
}
