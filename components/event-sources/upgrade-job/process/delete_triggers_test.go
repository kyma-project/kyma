package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteTriggers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Initialize fake client with triggers
	e2eSetup := newE2ESetup()

	p := &Process{
		Logger: logrus.New(),
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Delete triggers", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		deleteTriggers := NewDeleteTriggers(p)
		p.Steps = []Step{
			saveCurrentState,
			deleteTriggers,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check for deleted triggers
		for _, trigger := range e2eSetup.triggers.Items {
			_, err := p.Clients.Trigger.Get(trigger.Namespace, trigger.Name)
			g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
		}
	})
}
