package process

import (
	"log"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/onsi/gomega"
)

// TestCheckIsBebEnabled tests the CheckIsBebEnabled_DO step
func TestCheckIsBebEnabled(t *testing.T) {
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

	t.Run("Check is beb enabled", func(t *testing.T) {
		//// CASE 1: BEB
		// First check if the initial value is false for IsBebEnabled
		g.Expect(p.State.IsBebEnabled).Should(gomega.BeFalse())

		// @TODO: Fix this test

		//// Now check is beb enabled
		//p.Steps = []Step{
		//	NewCheckIsBebEnabled(p),
		//}
		//err := p.Execute()
		//g.Expect(err).Should(gomega.BeNil())
		//
		//// Check if the IsBebEnabled value is true
		//g.Expect(p.State.IsBebEnabled).Should(gomega.BeTrue())
		//
		////// CASE 1: NATS
		//// Now, change the Backend type to NATS and test again
		//e2eSetup.eventingBackends.Items[0].Status.Backend = eventingv1alpha1.NatsBackendType
		//p.Clients = getProcessClients(e2eSetup, g)
		//
		//// Now check is beb enabled
		//err = p.Execute()
		//g.Expect(err).Should(gomega.BeNil())
		//
		//// Check if the IsBebEnabled value is true
		//g.Expect(p.State.IsBebEnabled).Should(gomega.BeFalse())
	})
}
