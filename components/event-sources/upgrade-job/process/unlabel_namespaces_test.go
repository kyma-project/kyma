package process

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestUnlabelNamespaces(t *testing.T) {
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

	t.Run("Unlabel namespace", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		unLabelNs := NewUnLabelNamespace(p)
		p.Steps = []Step{
			saveCurrentState,
			unLabelNs,
		}
		expectedLabels := map[string]string{
			"foo": "should-not-be-removed",
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())
		gotNamespaces, err := p.Clients.Namespace.List()
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(len(gotNamespaces.Items)).To(gomega.Equal(len(e2eSetup.namespaces.Items)))
		for _, expectedNs := range e2eSetup.namespaces.Items {
			for _, gotNs := range gotNamespaces.Items {
				if gotNs.Name == expectedNs.Name {
					g.Expect(gotNs.Labels).To(gomega.Equal(expectedLabels))
					g.Expect(gotNs.Labels[knativeEventingLabelKey]).To(gomega.BeEmpty())
				}
			}
		}
	})
}
