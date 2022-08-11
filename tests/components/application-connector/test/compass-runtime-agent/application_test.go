package compass_runtime_agent

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestApplicationCrdCompare(t *testing.T) {

	services := make([]v1alpha1.Service, 0, 0)
	entries := make([]v1alpha1.Entry, 0, 0)

	comparer := Secret{}

	credentials := v1alpha1.Credentials{
		Type:              "OAuth",
		SecretName:        "secretTest", //TODO secret comaparsion, other application -> pass to other "klocek"!
		AuthenticationUrl: "authURL",
		CSRFInfo:          &v1alpha1.CSRFInfo{TokenEndpointURL: "csrfTokenURL"},
	}

	entries = append(entries, v1alpha1.Entry{
		Type:                        "api",
		TargetUrl:                   "targetURL",
		SpecificationUrl:            "specURL",
		ApiType:                     "v1",
		Credentials:                 credentials,
		RequestParametersSecretName: "paramSecret",
		Name:                        "test2",
		ID:                          "t2",
		CentralGatewayUrl:           "centralURL",
		AccessLabel:                 "", //ignore for now
		GatewayUrl:                  "",
	})

	entries = append(entries, v1alpha1.Entry{
		Type:                        "api",
		TargetUrl:                   "targetURL",
		SpecificationUrl:            "specURL",
		ApiType:                     "v1",
		Credentials:                 credentials,
		RequestParametersSecretName: "paramSecret",
		Name:                        "test1",
		ID:                          "t1",
		CentralGatewayUrl:           "centralURL",
		AccessLabel:                 "",
		GatewayUrl:                  "",
	})

	services = append(services, v1alpha1.Service{
		ID:                        "serviceTest",
		Identifier:                "st1",
		Name:                      "srvTest1",
		DisplayName:               "srvTest1",
		Description:               "srvTest1",
		Entries:                   entries,
		AuthCreateParameterSchema: nil,
		Labels:                    nil,
		LongDescription:           "",
		ProviderDisplayName:       "",
		Tags:                      nil,
	})

	services = append(services, v1alpha1.Service{
		ID:                        "serviceTest2",
		Identifier:                "st2",
		Name:                      "srvTest2",
		DisplayName:               "srvTest2",
		Description:               "srvTest2",
		Entries:                   entries,
		AuthCreateParameterSchema: nil,
		Labels:                    nil,
		LongDescription:           "",
		ProviderDisplayName:       "",
		Tags:                      nil,
	})

	{
		applicationCRD := &v1alpha1.Application{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "test",
				SkipInstallation: false,
				Services:         services,
				Labels:           nil,
				Tenant:           "test",
				Group:            "test",
				CompassMetadata: &v1alpha1.CompassMetadata{
					ApplicationID:  "compassID1",
					Authentication: v1alpha1.Authentication{ClientIds: []string{"11", "22"}},
				},
				Tags:                []string{"tag1", "tag2"},
				DisplayName:         "applicationOneDisplay",
				ProviderDisplayName: "applicationOneDisplay",
				LongDescription:     "applicationOne Test",
				SkipVerify:          true,
			},
		}

		applicationCRD_bad := &v1alpha1.Application{
			TypeMeta:   v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "test",
				SkipInstallation: false,
				Services:         services,
				Labels:           nil,
				Tenant:           "test",
				Group:            "test",
				CompassMetadata: &v1alpha1.CompassMetadata{
					ApplicationID:  "compassID1",
					Authentication: v1alpha1.Authentication{ClientIds: []string{"11", "22"}},
				},
				Tags:                []string{"tag1", "tag2"},
				DisplayName:         "applicationOneDisplay",
				ProviderDisplayName: "applicationOneDisplay",
				LongDescription:     "applicationOne Test",
				SkipVerify:          true,
			},
			Status: v1alpha1.ApplicationStatus{},
		}

		testCases := []struct {
			testMessage string
			application *v1alpha1.Application
			result      bool
		}{
			{
				testMessage: "should return true when Application CRDs are equal",
				application: applicationCRD,
				result:      true,
			},
			{
				testMessage: "should return false when Application CRDs are not equal",
				application: applicationCRD_bad,
				result:      false,
			},
		}

		for _, test := range testCases {
			t.Run(test.testMessage, func(t *testing.T) {
				assert.Equal(t, test.result, Compare(test.application, applicationCRD, comparer))
			})
		}
	}
}
