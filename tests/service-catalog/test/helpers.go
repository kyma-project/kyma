package test

import (
	"fmt"
	"net/http"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func isCatalogForbidden(url string) (bool, int, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = fmt.Sprintf("%s/cluster", url)

	client, err := osb.NewClient(config)
	if err != nil {
		return false, 0, errors.Wrapf(err, "while creating osb client for broker with URL: %s", url)
	}
	var statusCode int
	isForbiddenError := func(err error) bool {
		statusCodeError, ok := err.(osb.HTTPStatusCodeError)
		if !ok {
			return false
		}
		statusCode = statusCodeError.StatusCode
		return statusCodeError.StatusCode == http.StatusForbidden
	}

	_, err = client.GetCatalog()
	switch {
	case err == nil:
		return false, http.StatusOK, nil
	case isForbiddenError(err):
		return true, statusCode, nil
	default:
		return false, statusCode, errors.Wrapf(err, "while getting catalog from broker with URL: %s", url)
	}
}

func fixNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}

func fixApplication(name string) *appTypes.Application {
	return &appTypes.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appTypes.ApplicationSpec{
			Description:      "Application used by acceptance test",
			AccessLabel:      "fix-access",
			SkipInstallation: true,
			Services: []appTypes.Service{
				{
					ID:   "id-00000-1234-test",
					Name: "provider-4951",
					Labels: map[string]string{
						"connected-app": name,
					},
					ProviderDisplayName: "provider",
					DisplayName:         name,
					Description:         "Application Service Class used by acceptance test",
					Tags:                []string{},
					Entries: []appTypes.Entry{
						{
							Type:        "API",
							AccessLabel: "acc-label",
							GatewayUrl:  "http://promotions-gateway.production.svc.cluster.local/",
						},
					},
				},
			},
		},
	}
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
