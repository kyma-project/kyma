package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteEventSources(t *testing.T) {
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

	t.Run("Delete event sources", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		deleteEventSources := NewDeleteEventSources(p)
		p.Steps = []Step{
			saveCurrentState,
			deleteEventSources,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check for event sources for each app
		for _, app := range e2eSetup.applications.Items {
			_, err := p.Clients.HttpSource.Get(KymaIntegrationNamespace, app.Name)
			g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
		}
	})
}
