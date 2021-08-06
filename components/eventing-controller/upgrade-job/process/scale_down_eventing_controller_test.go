package process

import (
	"log"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/processtest"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/onsi/gomega"
)

// TestScaleDownEventingController tests the ScaleDownEventingController_DO step
func TestScaleDownEventingController(t *testing.T) {
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

	t.Run("Scale down eventing controller deployment", func(t *testing.T) {
		// First check if eventing-controller deployment has replicas=1
		expectedDeploymentName := "eventing-controller"
		ec, err := p.Clients.Deployment.Get(processtest.KymaSystemNamespace, expectedDeploymentName)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(*ec.Spec.Replicas).Should(gomega.Equal(int32(1)))

		// Now scale down eventing controller deployment
		p.Steps = []Step{
			NewScaleDownEventingController(p),
		}
		err = p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if if eventing-controller deployment has replicas=0
		ec, err = p.Clients.Deployment.Get(processtest.KymaSystemNamespace, expectedDeploymentName)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(*ec.Spec.Replicas).Should(gomega.Equal(int32(0)))
	})
}
