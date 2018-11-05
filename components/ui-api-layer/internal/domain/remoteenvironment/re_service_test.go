package remoteenvironment_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func TestServiceListNamespacesForRemoteEnvironmentSuccess(t *testing.T) {
	// given
	const fixMappingName = "test-re"
	fixMapping := fixEnvironmentMappingCR(fixMappingName, "production")

	client := fake.NewSimpleClientset(&fixMapping)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	reSharedInformers := informerFactory.Applicationconnector().V1alpha1()
	emInformer := reSharedInformers.EnvironmentMappings().Informer()
	reInformer := reSharedInformers.RemoteEnvironments().Informer()

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, emInformer, nil, reInformer)
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
	reSharedInformers := informerFactory.Applicationconnector().V1alpha1()
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
	reSharedInformers := informerFactory.Applicationconnector().V1alpha1()
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
	reSharedInformers := informerFactory.Applicationconnector().V1alpha1()
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
	reSharedInformers := informerFactory.Applicationconnector().V1alpha1()

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

func TestGetConnectionURLSuccess(t *testing.T) {
	// given
	testServer := newTestServer(`{"url": "http://gotURL-with-token", "token": "token"}`, http.StatusCreated)
	defer testServer.Close()

	config := remoteenvironment.Config{
		Connector: remoteenvironment.ConnectorSvcCfg{
			URL: testServer.URL,
		},
	}

	svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, config, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)
	// when
	gotURL, err := svc.GetConnectionURL("fixRemoteEnvironmentName")

	// then
	require.NoError(t, err)
	assert.Equal(t, "http://gotURL-with-token", gotURL)
}

func TestGetConnectionURLFailure(t *testing.T) {
	t.Run("Should return an error in case of improper remote environment name", func(t *testing.T) {
		// given
		cfg := remoteenvironment.Config{
			Connector: remoteenvironment.ConnectorSvcCfg{
				URL: "connectorURL",
			},
		}

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, cfg, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("invalid/RemoteEnvironmentName")

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

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, config, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("fixRemoteEnvironment")

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

		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, cfg, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("fixRemoteEnvironment")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `while extracting connection URL from body: while decoding json: invalid character 's' looking for beginning of value`)
	})
}

func TestRemoteEnvironmentService_Create(t *testing.T) {
	// GIVEN
	client := fake.NewSimpleClientset()
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)

	// WHEN
	re, err := svc.Create(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, re.Name, fixName)
	assert.Equal(t, re.Spec.Description, fixDesc)
	assert.Equal(t, re.Spec.Labels, fixLabels)
}

func TestRemoteEnvironmentService_Delete(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)

	// WHEN
	err = svc.Delete(fixName)

	// THEN
	require.NoError(t, err)
	_, err = client.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(fixName, v1.GetOptions{})
	assert.True(t, apiErrors.IsNotFound(err))
}

func TestRemoteEnvironmentService_Update(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))
	informerFactory := externalversions.NewSharedInformerFactory(client, time.Second)
	informer := informerFactory.Applicationconnector().V1alpha1().RemoteEnvironments().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, informer)
	require.NoError(t, err)

	// WHEN
	re, err := svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixLabels, re.Spec.Labels)
	assert.Equal(t, fixDesc, re.Spec.Description)
}

func TestRemoteEnvironmentService_Update_ErrorInRetryLoop(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))
	client.PrependReactor("update", "remoteenvironments", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("fix")
	})
	informerFactory := externalversions.NewSharedInformerFactory(client, time.Second)
	informer := informerFactory.Applicationconnector().V1alpha1().RemoteEnvironments().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, informer)
	require.NoError(t, err)

	// WHEN
	_, err = svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while updating %s [%s]: fix", pretty.RemoteEnvironment, fixName))
}

func TestRemoteEnvironmentService_Update_SuccessAfterRetry(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))
	i := 0
	client.PrependReactor("update", "remoteenvironments", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		if i < 3 {
			i++
			return true, nil, apiErrors.NewConflict(schema.GroupResource{}, "", errors.New("fix"))
		}
		return false, fixRemoteEnvironmentCR(fixName), nil
	})
	informerFactory := externalversions.NewSharedInformerFactory(client, time.Second)
	informer := informerFactory.Applicationconnector().V1alpha1().RemoteEnvironments().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, informer)
	require.NoError(t, err)

	// WHEN
	re, err := svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixLabels, re.Spec.Labels)
	assert.Equal(t, fixDesc, re.Spec.Description)
}

func TestRemoteEnvironmentService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)
		svc.Subscribe(remoteEnvironmentListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		svc.Subscribe(remoteEnvironmentListener)
		svc.Subscribe(remoteEnvironmentListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListenerA := listener.NewRemoteEnvironment(nil, nil)
		remoteEnvironmentListenerB := listener.NewRemoteEnvironment(nil, nil)

		svc.Subscribe(remoteEnvironmentListenerA)
		svc.Subscribe(remoteEnvironmentListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestRemoteEnvironmentService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)
		svc.Subscribe(remoteEnvironmentListener)

		svc.Unsubscribe(remoteEnvironmentListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)
		svc.Subscribe(remoteEnvironmentListener)
		svc.Subscribe(remoteEnvironmentListener)

		svc.Unsubscribe(remoteEnvironmentListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		remoteEnvironmentListenerA := listener.NewRemoteEnvironment(nil, nil)
		remoteEnvironmentListenerB := listener.NewRemoteEnvironment(nil, nil)
		svc.Subscribe(remoteEnvironmentListenerA)
		svc.Subscribe(remoteEnvironmentListenerB)

		svc.Unsubscribe(remoteEnvironmentListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := remoteenvironment.NewRemoteEnvironmentService(nil, remoteenvironment.Config{}, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		svc.Unsubscribe(nil)
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
