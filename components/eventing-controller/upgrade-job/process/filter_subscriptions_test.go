package process

import (
	"log"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/onsi/gomega"
)

// TestFilterSubscriptions tests the FilterSubscriptions_DO step
func TestFilterSubscriptions(t *testing.T) {
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
		Logger:             ctrLogger.Logger,
		TimeoutPeriod:      60 * time.Second,
		ReleaseName:        cfg.ReleaseName,
		Domain:             cfg.Domain,
		KymaNamespace:      cfg.KymaNamespace,
		ControllerName:     cfg.EventingControllerName,
		PublisherName:      "eventing-publisher-proxy",
		PublisherNamespace: cfg.KymaNamespace,
		State:              State{},
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Filter subscriptions", func(t *testing.T) {
		// Run the job
		p.Steps = []Step{
			NewGetSubscriptions(p),
			NewFilterSubscriptions(p),
		}
		err = p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if total subscriptions are 4
		g.Expect(p.State.Subscriptions).Should(gomega.Not(gomega.BeNil()))
		g.Expect(p.State.Subscriptions.Items).Should(gomega.HaveLen(4))

		// FilteredSubscriptions should have length 3 because 1 subscription was marked
		// as migrated in our dummy data
		g.Expect(p.State.FilteredSubscriptions).Should(gomega.Not(gomega.BeNil()))
		g.Expect(p.State.FilteredSubscriptions.Items).Should(gomega.HaveLen(3))
	})
}
