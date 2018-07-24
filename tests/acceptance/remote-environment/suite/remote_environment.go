package suite

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) createRemoteEnvironmentResources() {
	client, err := versioned.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	displayName := fmt.Sprintf("acc-test-re-name-%s", ts.TestID)
	rei := client.RemoteenvironmentV1alpha1().RemoteEnvironments(ts.namespace)
	_, err = rei.Create(fixRemoteEnvironment(ts.remoteEnvironmentName, ts.accessLabel, ts.osbServiceId, ts.gatewayUrl, displayName))
	require.NoError(ts.t, err)

	emi := client.RemoteenvironmentV1alpha1().EnvironmentMappings(ts.namespace)
	_, err = emi.Create(fixEnvironmentMapping(ts.remoteEnvironmentName))
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteRemoteEnvironment() {
	client, err := versioned.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	rei := client.RemoteenvironmentV1alpha1().RemoteEnvironments(ts.namespace)
	err = rei.Delete(ts.remoteEnvironmentName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func fixEnvironmentMapping(name string) *v1alpha1.EnvironmentMapping {
	return &v1alpha1.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "remoteenvironment.kyma.cx/v1alpha1",
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
			APIVersion: "remoteenvironment.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Source: v1alpha1.Source{
				Namespace:   "com.ns",
				Type:        "commerce",
				Environment: "production",
			},
			AccessLabel: "re-access-label",
			Description: "Remote Environment used by remote-environment acceptance test",
			Services: []v1alpha1.Service{
				{
					ID:                  serviceId,
					ProviderDisplayName: "provider",
					DisplayName:         displayName,
					LongDescription:     "Remote Environment Service Class used by remote-environment acceptance test",
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
