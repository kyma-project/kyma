package k8s_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
)

func TestResourceService_Create(t *testing.T) {
	const (
		namespace  = "test-namespace"
		name       = "test-name"
		pluralName = "pods"
		kind       = "Pod"
	)
	var testCases = []struct {
		caseName          string
		namespace         string
		apiVersion        string
		failingPluralName bool
		failingRESTClient bool
		success           bool
	}{
		{"Success", "test-namespace", "v1", false, false, true},
		{"WithAPIGroup", "test-namespace", "test/v1beta1", false, false, true},
		{"NamespaceMissmatch", "test-missmatch", "v1", false, false, false},
		{"PluralNotFound", "test-namespace", "v1", true, false, false},
		{"ErrorCreatingResource", "test-namespace", "v1", false, true, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			resource, err := fixResource(name, testCase.namespace, kind, testCase.apiVersion)
			require.NoError(t, err)
			client := fixClient(testCase.apiVersion, kind, pluralName, resource, testCase.failingPluralName, testCase.failingRESTClient)
			svc := k8s.NewResourceService(client)

			result, err := svc.Create(namespace, resource)

			if testCase.success {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
				assert.Nil(t, result)
			}

		})
	}
}

func fixResource(name, namespace, kind, apiVersion string) (types.Resource, error) {
	resourceJSON := gqlschema.JSON{
		"kind":       kind,
		"apiVersion": apiVersion,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}

	converter := k8s.NewResourceConverter()
	return converter.GQLJSONToResource(resourceJSON)
}

func fixClient(apiVersion, kind, pluralName string, created types.Resource, failingPluralName, failingRESTClient bool) discovery.DiscoveryInterface {
	fakeResources := &v1.APIResourceList{
		TypeMeta:     v1.TypeMeta{},
		GroupVersion: apiVersion,
	}
	if !failingPluralName {
		fakeResources.APIResources = []v1.APIResource{
			v1.APIResource{
				Name: pluralName,
				Kind: kind,
			},
		}
	}

	client := automock.DiscoveryInterface{}
	fakeRESTClient := &fake.RESTClient{
		NegotiatedSerializer: scheme.Codecs,
		Client:               fixHTTPClient(created.Body, failingRESTClient),
	}

	client.On("ServerResourcesForGroupVersion", apiVersion).Return(fakeResources, nil).Once()
	client.On("RESTClient").Return(fakeRESTClient, nil).Once()

	return &client
}

func fixHTTPClient(body []byte, failing bool) *http.Client {
	var res http.Response
	var err error

	if failing {
		res, err = http.Response{StatusCode: 200, Header: defaultHeader(), Body: nil}, errors.New("error")
	} else {
		res, err = http.Response{StatusCode: 200, Header: defaultHeader(), Body: ioutil.NopCloser(bytes.NewReader(body))}, nil
	}

	return fake.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
		return &res, err
	})
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}
