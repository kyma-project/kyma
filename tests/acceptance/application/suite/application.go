package suite

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingClientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appClientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) createApplicationResources() {
	aClientset, err := appClientset.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	displayName := fmt.Sprintf("acc-test-app-name-%s", ts.TestID)
	_, err = createApplication(aClientset.ApplicationconnectorV1alpha1().Applications(), ts.applicationName, ts.accessLabel, ts.osbServiceId, ts.gatewayUrl, displayName)
	require.NoError(ts.t, err)

	mClientset, err := mappingClientset.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	_, err = createApplicationMapping(mClientset.ApplicationconnectorV1alpha1().ApplicationMappings(ts.namespace), ts.applicationName)
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteApplication() {
	client, err := appClientset.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	err = client.ApplicationconnectorV1alpha1().Applications().Delete(ts.applicationName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func createApplication(rei appCli.ApplicationInterface, appName, accessLabel, serviceId, gatewayUrl, displayName string) (*appTypes.Application, error) {
	return rei.Create(fixApplication(appName, accessLabel, serviceId, gatewayUrl, displayName))
}

func createApplicationMapping(emi mappingCli.ApplicationMappingInterface, appName string) (*mappingTypes.ApplicationMapping, error) {
	return emi.Create(fixApplicationMapping(appName))
}

func fixApplicationMapping(name string) *mappingTypes.ApplicationMapping {
	return &mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixApplication(name, accessLabel, serviceId, gatewayUrl, displayName string) *appTypes.Application {
	return &appTypes.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appTypes.ApplicationSpec{
			AccessLabel: "re-access-label",
			Description: "Application used by application acceptance test",
			Services: []appTypes.Service{
				{
					ID:   serviceId,
					Name: serviceId,
					Labels: map[string]string{
						"connected-app": name,
					},
					ProviderDisplayName: "provider",
					DisplayName:         displayName,
					Description:         "Application Service Class used by application acceptance test",
					Tags:                []string{},
					Entries: []appTypes.Entry{
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
