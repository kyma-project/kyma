package process

import (
	"log"
	"testing"
	"time"

	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/processtest"
)

// TestDeletePublisherDeployment tests the DeletePublisherDeployment_DO step
func TestDeletePublisherDeployment(t *testing.T) {
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

	t.Run("Delete eventing publisher proxy deployment", func(t *testing.T) {
		// First check if eventing-publisher-proxy deployment exists
		expectedDeploymentName := "eventing-publisher-proxy"
		_, err = p.Clients.Deployment.Get(processtest.KymaSystemNamespace, expectedDeploymentName)
		g.Expect(err).Should(gomega.BeNil())

		// Now delete the eventing-publisher-proxy deployment
		p.Steps = []Step{
			NewDeletePublisherDeployment(p),
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if eventing-publisher-proxy deployment is deleted
		_, err = p.Clients.Deployment.Get(processtest.KymaSystemNamespace, expectedDeploymentName)
		g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
	})
}
