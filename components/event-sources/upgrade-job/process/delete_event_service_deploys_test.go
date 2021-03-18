package process

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteEventServiceDeploys(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	e2eSetup := newE2ESetup()

	p := &Process{
		Steps:           nil,
		ReleaseName:     "release",
		BEBNamespace:    "beb-ns",
		EventingBackend: "beb",
		EventTypePrefix: "prefix",
		Clients:         Clients{},
		Logger:          logrus.New(),
		State:           State{},
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Delete event service deployments", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		deleteEventServiceDeploys := NewDeleteEventServiceDeployments(p)
		p.Steps = []Step{
			saveCurrentState,
			deleteEventServiceDeploys,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check for event service deployments for each app
		for _, app := range e2eSetup.applications.Items {
			expectedDeploymentName := fmt.Sprintf("%s-event-service", app.Name)
			_, err := p.Clients.Deployment.Get(KymaIntegrationNamespace, expectedDeploymentName)
			g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
		}
	})
}
