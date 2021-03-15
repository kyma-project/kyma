package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestCleanup(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Initializing fake client with brokers
	e2eSetup := newE2ESetup()

	p := &Process{
		ReleaseName: "release",
		Logger:      logrus.New(),
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Cleanup", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		cleanUp := NewCleanUp(p)
		p.Steps = []Step{
			saveCurrentState,
			cleanUp,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check for deleted config map
		_, err = p.Clients.ConfigMap.Get(BackedUpConfigMapNamespace, BackedUpConfigMapName)
		g.Expect(k8serrors.IsNotFound(err)).To(gomega.BeTrue())
	})
}
