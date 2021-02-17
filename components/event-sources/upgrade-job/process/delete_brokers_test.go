package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteBrokers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Initializing fake client with brokers
	e2eSetup := newE2ESetup()

	p := &Process{
		ReleaseName: "release",
		Logger:      logrus.New(),
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Delete brokers", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		deleteBrokers := NewDeleteBrokers(p)
		p.Steps = []Step{
			saveCurrentState,
			deleteBrokers,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check for deleted brokers
		for _, broker := range e2eSetup.brokers.Items {
			_, err := p.Clients.Broker.Get(broker.Namespace, broker.Name)
			g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
		}
	})
}
