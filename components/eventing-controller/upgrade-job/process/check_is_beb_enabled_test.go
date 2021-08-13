package process

import (
	"log"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
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
		Logger:             ctrLogger.Logger,
		TimeoutPeriod:      60 * time.Second,
		ReleaseName:        cfg.ReleaseName,
		Domain:             cfg.Domain,
		KymaNamespace:      cfg.KymaNamespace,
		ControllerName:     cfg.EventingControllerName,
		PublisherName:      "eventing-publisher-proxy",
		PublisherNamespace: cfg.KymaNamespace,
		State: State{
			Is124Cluster: true,
		},
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Check is beb enabled", func(t *testing.T) {
		//// CASE 1: BEB
		// First check if the initial value is false for IsBebEnabled
		g.Expect(p.State.IsBebEnabled).Should(gomega.BeFalse())

		// Now check is beb enabled
		p.Steps = []Step{
			NewCheckIsBebEnabled(p),
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if the IsBebEnabled value is true
		g.Expect(p.State.IsBebEnabled).Should(gomega.BeTrue())

		//// CASE 2: NATS
		// Now, change the Backend type to NATS and test again
		e2eSetup.eventingBackends.Items[0].Status.Backend = eventingv1alpha1.NatsBackendType
		e2eSetup.secrets = &corev1.SecretList{} // Remove BEB secret

		p.Clients = getProcessClients(e2eSetup, g)

		// Now check is beb enabled
		err = p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if the IsBebEnabled value is true
		g.Expect(p.State.IsBebEnabled).Should(gomega.BeFalse())
	})
}
