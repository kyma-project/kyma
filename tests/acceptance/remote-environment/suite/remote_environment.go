package suite

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	v1alpha12 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) createRemoteEnvironmentResources() {
	client, err := versioned.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	displayName := fmt.Sprintf("acc-test-re-name-%s", ts.TestID)
	rei := client.ApplicationconnectorV1alpha1().RemoteEnvironments()
	_, err = createRemoteEnvironment(rei, ts.remoteEnvironmentName, ts.accessLabel, ts.osbServiceId, ts.gatewayUrl, displayName)
	require.NoError(ts.t, err)

	emi := client.ApplicationconnectorV1alpha1().EnvironmentMappings(ts.namespace)
	_, err = createEnvironmentMapping(emi, ts.remoteEnvironmentName)
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteRemoteEnvironment() {
	client, err := versioned.NewForConfig(ts.config)
	require.NoError(ts.t, err)
	rei := client.ApplicationconnectorV1alpha1().RemoteEnvironments()

	err = rei.Delete(ts.remoteEnvironmentName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func createRemoteEnvironment(rei v1alpha12.RemoteEnvironmentInterface, reName, accessLabel, serviceId, gatewayUrl, displayName string) (*v1alpha1.RemoteEnvironment, error) {
	return rei.Create(fixRemoteEnvironment(reName, accessLabel, serviceId, gatewayUrl, displayName))
}

func createEnvironmentMapping(emi v1alpha12.EnvironmentMappingInterface, reName string) (*v1alpha1.EnvironmentMapping, error) {
	return emi.Create(fixEnvironmentMapping(reName))
}

func fixEnvironmentMapping(name string) *v1alpha1.EnvironmentMapping {
	return &v1alpha1.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixRemoteEnvironment(name, accessLabel, serviceId, gatewayUrl, displayName string) *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RemoteEnvironment",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			AccessLabel: "re-access-label",
			Description: "Remote Environment used by remote-environment acceptance test",
			Services: []v1alpha1.Service{
				{
					ID:   serviceId,
					Name: serviceId,
					Labels: map[string]string{
						"connected-app": name,
					},
					ProviderDisplayName: "provider",
					DisplayName:         displayName,
					Description:         "Remote Environment Service Class used by remote-environment acceptance test",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
						{
							Type:        "API",
							AccessLabel: accessLabel,
							GatewayUrl:  gatewayUrl,
						},
					},
				},
			},
		},
	}
}
