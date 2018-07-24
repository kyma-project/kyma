package remoteenvironment_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestServiceListNamespacesForRemoteEnvironmentSuccess(t *testing.T) {
	// given
	const fixMappingName = "test-re"
	fixMapping := fixEnvironmentMappingCR(fixMappingName, "production")

	client := fake.NewSimpleClientset(&fixMapping)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Remoteenvironment().V1alpha1()
	emInformer := reSharedInformers.EnvironmentMappings().Informer()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, emInformer, nil, nil)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, emInformer)

	// when
	nsList, err := svc.ListNamespacesFor(fixMappingName)

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 1)
	assert.Equal(t, nsList[0], fixMapping.Namespace)
}

func TestServiceFindRemoteEnvironmentSuccess(t *testing.T) {
	// given
	reName := "testExample"

	fixRemoteEnv := fixRemoteEnvironmentCR("testExample")
	client := fake.NewSimpleClientset(fixRemoteEnv)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Remoteenvironment().V1alpha1()
	reInformer := reSharedInformers.RemoteEnvironments().Informer()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, reSharedInformers.EnvironmentMappings().Informer(), nil, reInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, reInformer)

	// when
	re, err := svc.Find(reName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixRemoteEnv, re)
}

func TestServiceFindRemoteEnvironmentFail(t *testing.T) {
	// given
	reName := "testExample"

	client := fake.NewSimpleClientset()

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Remoteenvironment().V1alpha1()
	reInformer := reSharedInformers.RemoteEnvironments().Informer()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, reSharedInformers.EnvironmentMappings().Informer(), nil, reInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, reInformer)

	// when
	re, err := svc.Find(reName)

	// then
	require.NoError(t, err)
	assert.Nil(t, re)
}

func TestServiceListAllRemoteEnvironmentsSuccess(t *testing.T) {
	// given
	fixREA := fixRemoteEnvironmentsCR("re-name-a")
	fixREB := fixRemoteEnvironmentsCR("re-name-b")

	client := fake.NewSimpleClientset(&fixREA, &fixREB)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Remoteenvironment().V1alpha1()
	reInformer := reSharedInformers.RemoteEnvironments().Informer()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, reSharedInformers.EnvironmentMappings().Informer(), nil, reInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, reInformer)

	// when
	nsList, err := svc.List(pager.PagingParams{})

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 2)
	assert.Contains(t, nsList, &fixREB)
	assert.Contains(t, nsList, &fixREA)
}

func TestServiceListRemoteEnvironmentsInEnvironmentSuccess(t *testing.T) {
	// given
	const fixEnvironment = "prod"

	fixREA := fixRemoteEnvironmentsCR("re-name-a")
	fixREB := fixRemoteEnvironmentsCR("re-name-b")
	fixMappingREA := fixEnvironmentMappingCR("re-name-a", fixEnvironment)
	fixMappingREB := fixEnvironmentMappingCR("re-name-b", fixEnvironment)

	client := fake.NewSimpleClientset(&fixREA, &fixREB, &fixMappingREA, &fixMappingREB)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Remoteenvironment().V1alpha1()

	reInformer := reSharedInformers.RemoteEnvironments().Informer()
	emInformer := reSharedInformers.EnvironmentMappings().Informer()
	emLister := reSharedInformers.EnvironmentMappings().Lister()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, emInformer, emLister, reInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, reInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, emInformer)

	// when
	nsList, err := svc.ListInEnvironment(fixEnvironment)

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 2)
	assert.Contains(t, nsList, &fixREA)
	assert.Contains(t, nsList, &fixREB)
}

func TestGetConnectionUrlSuccess(t *testing.T) {
	// given
	testServer := newTestServer(`{"url": "http://gotURL-with-token", "token": "token"}`, http.StatusCreated)
	defer testServer.Close()

	config := remoteenvironment.Config{
		Connector: remoteenvironment.ConnectorSvcCfg{
			URL: testServer.URL,
		},
	}

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, config, newDummyInformer(), nil, nil)
	require.NoError(t, err)
	// when
	gotURL, err := svc.GetConnectionUrl("fixRemoteEnvironmentName")

	// then
	require.NoError(t, err)
	assert.Equal(t, "http://gotURL-with-token", gotURL)
}

func TestGetConnectionUrlFailure(t *testing.T) {
	t.Run("Should return an error in case of improper remote environment name", func(t *testing.T) {
		// given
		cfg := remoteenvironment.Config{
			Connector: remoteenvironment.ConnectorSvcCfg{
				URL: "connectorUrl",
			},
		}

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, cfg, newDummyInformer(), nil, nil)
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionUrl("invalid/RemoteEnvironmentName")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `Remote evironment name "invalid/RemoteEnvironmentName" does not match regex: ^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	})

	t.Run("Should return an error in case of 403 status code", func(t *testing.T) {
		// given
		testServer := newTestServer(`{"code": 403, "error": "Invalid token."}`, http.StatusForbidden)
		defer testServer.Close()

		config := remoteenvironment.Config{
			Connector: remoteenvironment.ConnectorSvcCfg{
				URL: testServer.URL,
			},
		}

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, config, newDummyInformer(), nil, nil)
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionUrl("fixRemoteEnvironment")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `while requesting connection URL obtained unexpected status code 403: Invalid token.`)
	})

	t.Run("Should return an error in case of invalid json format", func(t *testing.T) {
		// given
		testServer := newTestServer("something", http.StatusCreated)
		defer testServer.Close()

		cfg := remoteenvironment.Config{
			Connector: remoteenvironment.ConnectorSvcCfg{
				URL: testServer.URL,
			},
		}

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, cfg, newDummyInformer(), nil, nil)
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionUrl("fixRemoteEnvironment")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `while extracting connection URL from body: while decoding json: invalid character 's' looking for beginning of value`)
	})
}

func newDummyInformer() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(&cache.ListWatch{}, nil, 0, cache.Indexers{})
}

func newTestServer(data string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, data)
	}))
}

func fixEnvironmentMappingCR(name, ns string) v1alpha1.EnvironmentMapping {
	return v1alpha1.EnvironmentMapping{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

}

func fixRemoteEnvironmentsCR(name string) v1alpha1.RemoteEnvironment {
	return v1alpha1.RemoteEnvironment{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}

func fixRemoteEnvironmentCR(name string) *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}
