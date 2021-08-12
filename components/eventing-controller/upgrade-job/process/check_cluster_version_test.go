package process

import (
	"log"
	"testing"
	"time"

	"github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

// TestCheckClusterVersion tests the CheckClusterVersion_DO step
func TestCheckClusterVersion(t *testing.T) {
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

	t.Run("Check cluster version", func(t *testing.T) {
		//// CASE 1: 1.24.x (i.e. With Eventing Backend)
		// First check if the initial value is false for Is124Cluster
		g.Expect(p.State.Is124Cluster).Should(gomega.BeFalse())

		// Now check is beb enabled
		p.Steps = []Step{
			NewCheckClusterVersion(p),
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if the Is124Cluster value is true
		g.Expect(p.State.Is124Cluster).Should(gomega.BeTrue())

		//// CASE 2: 1.23.x (i.e. Without Eventing Backend)
		// Now, remove the eventing-backend instance
		e2eSetup.eventingBackends = &eventingv1alpha1.EventingBackendList{}

		p.Clients = getProcessClients(e2eSetup, g)

		// Now check is beb enabled
		err = p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if the IsBebEnabled value is true
		g.Expect(p.State.Is124Cluster).Should(gomega.BeFalse())
	})
}
