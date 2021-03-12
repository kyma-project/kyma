package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteAppSubscriptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Initialize fake client with app subscriptions
	e2eSetup := newE2ESetup()

	p := &Process{
		Logger: logrus.New(),
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Delete triggers", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		deleteAppSubscriptions := NewDeleteAppSubscriptions(p)
		p.Steps = []Step{
			saveCurrentState,
			deleteAppSubscriptions,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(len(e2eSetup.appSubscriptions.Items)).To(gomega.Equal(1))

		// Check for deleted subscriptions for application
		for _, oldAppSub := range e2eSetup.appSubscriptions.Items {
			if oldAppSub.Labels[applicationNameKey] != "" {
				_, err := p.Clients.Subscription.Get(KymaIntegrationNamespace, oldAppSub.Name)
				g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
			}
		}
	})
}
