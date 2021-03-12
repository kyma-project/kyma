package process

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestPatchConnectivityValidators(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Initialize fake client with connectivity validators
	e2eSetup := newE2ESetup()

	p := &Process{
		Logger: logrus.New(),
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Patch connectivity validators", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		patchConnectivityValidators := NewPatchConnectivityValidators(p)
		p.Steps = []Step{
			saveCurrentState,
			patchConnectivityValidators,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())
		// Check for deleted triggers
		for _, app := range e2eSetup.applications.Items {
			expectedDeploymentArgs := getNewContainerArgs(app.Name)
			deploymentName := fmt.Sprintf("%s-connectivity-validator", app.Name)
			containerName := deploymentName
			var container corev1.Container
			gotDeploy, err := p.Clients.Deployment.Get(KymaIntegrationNamespace, deploymentName)
			g.Expect(err).Should(gomega.BeNil())

			// Fetching the connectivity validator container
			for _, c := range gotDeploy.Spec.Template.Spec.Containers {
				if c.Name == containerName {
					container = c
					break
				}
			}
			g.Expect(container.Args).To(gomega.Equal(expectedDeploymentArgs))
		}
	})
}
