package application_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingFakeCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	mappingInformer "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appFakeCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/fake"
	appInformer "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/pretty"
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

func TestServiceListNamespacesForApplicationSuccess(t *testing.T) {
	// given
	const fixMappingName = "test-mapping"
	fixMapping := fixApplicationMappingCR(fixMappingName, "production")

	// Mapping
	mCli := mappingFakeCli.NewSimpleClientset(&fixMapping)

	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, 0)
	mInformer := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Informer()

	// Application
	aCli := appFakeCli.NewSimpleClientset()

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	svc, err := application.NewApplicationService(application.Config{}, nil, nil, mInformer, nil, aInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, mInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	// when
	nsList, err := svc.ListNamespacesFor(fixMappingName)

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 1)
	assert.Equal(t, nsList[0], fixMapping.Namespace)
}

func TestServiceFindApplicationSuccess(t *testing.T) {
	// given
	appName := "testExample"

	fixApp := fixApplicationCR("testExample")

	// Mapping
	mCli := mappingFakeCli.NewSimpleClientset()

	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, 0)
	mInformer := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Informer()

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixApp)

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	svc, err := application.NewApplicationService(application.Config{}, nil, nil, mInformer, nil, aInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, mInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	// when
	app, err := svc.Find(appName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixApp, app)
}

func TestServiceFindApplicationFail(t *testing.T) {
	// given
	appName := "testExample"

	// Mapping
	mCli := mappingFakeCli.NewSimpleClientset()

	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, 0)
	mInformer := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Informer()

	// Application
	aCli := appFakeCli.NewSimpleClientset()

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	svc, err := application.NewApplicationService(application.Config{}, nil, nil, mInformer, nil, aInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, mInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	// when
	app, err := svc.Find(appName)

	// then
	require.NoError(t, err)
	assert.Nil(t, app)
}

func TestServiceListAllApplicationsSuccess(t *testing.T) {
	// given
	fixAppA := fixApplicationCR("app-name-a")
	fixAppB := fixApplicationCR("app-name-b")

	// Mapping
	mCli := mappingFakeCli.NewSimpleClientset()

	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, 0)
	mInformer := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Informer()

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixAppA, fixAppB)

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	svc, err := application.NewApplicationService(application.Config{}, nil, nil, mInformer, nil, aInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, mInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	// when
	nsList, err := svc.List(pager.PagingParams{})

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 2)
	assert.Contains(t, nsList, fixAppB)
	assert.Contains(t, nsList, fixAppA)
}

func TestServiceListApplicationsInEnvironmentSuccess(t *testing.T) {
	// given
	const fixEnvironment = "prod"

	fixAppA := fixApplicationCR("app-name-a")
	fixAppB := fixApplicationCR("app-name-b")
	fixMappingAppA := fixApplicationMappingCR("app-name-a", fixEnvironment)
	fixMappingAppB := fixApplicationMappingCR("app-name-b", fixEnvironment)

	// Mapping
	mCli := mappingFakeCli.NewSimpleClientset(&fixMappingAppA, &fixMappingAppB)

	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, 0)
	mInformer := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Informer()
	mLister := mInformerFactory.Applicationconnector().V1alpha1().ApplicationMappings().Lister()

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixAppA, fixAppB)

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	svc, err := application.NewApplicationService(application.Config{}, nil, nil, mInformer, mLister, aInformer)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, mInformer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	// when
	nsList, err := svc.ListInEnvironment(fixEnvironment)

	// then
	require.NoError(t, err)

	require.Len(t, nsList, 2)
	assert.Contains(t, nsList, fixAppA)
	assert.Contains(t, nsList, fixAppB)
}

func TestGetConnectionURLSuccess(t *testing.T) {
	// given
	testServer := newTestServer(`{"url": "http://gotURL-with-token", "token": "token"}`, http.StatusCreated)
	defer testServer.Close()

	cfg := application.Config{
		Connector: application.ConnectorSvcCfg{
			URL: testServer.URL,
		},
	}

	svc, err := application.NewApplicationService(cfg, nil, nil, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)
	// when
	gotURL, err := svc.GetConnectionURL("fixApplicationName")

	// then
	require.NoError(t, err)
	assert.Equal(t, "http://gotURL-with-token", gotURL)
}

func TestGetConnectionURLFailure(t *testing.T) {
	t.Run("Should return an error in case of improper application name", func(t *testing.T) {
		// given
		cfg := application.Config{
			Connector: application.ConnectorSvcCfg{
				URL: "connectorURL",
			},
		}

		svc, err := application.NewApplicationService(cfg, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("invalid/ApplicationName")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `Application name "invalid/ApplicationName" does not match regex: ^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	})

	t.Run("Should return an error in case of 403 status code", func(t *testing.T) {
		// given
		testServer := newTestServer(`{"code": 403, "error": "Invalid token."}`, http.StatusForbidden)
		defer testServer.Close()

		cfg := application.Config{
			Connector: application.ConnectorSvcCfg{
				URL: testServer.URL,
			},
		}

		svc, err := application.NewApplicationService(cfg, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("fixApplication")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `while requesting connection URL obtained unexpected status code 403: Invalid token.`)
	})

	t.Run("Should return an error in case of invalid json format", func(t *testing.T) {
		// given
		testServer := newTestServer("something", http.StatusCreated)
		defer testServer.Close()

		cfg := application.Config{
			Connector: application.ConnectorSvcCfg{
				URL: testServer.URL,
			},
		}

		svc, err := application.NewApplicationService(cfg, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		// when
		gotURL, err := svc.GetConnectionURL("fixApplication")

		// then
		require.Error(t, err)
		assert.Empty(t, gotURL)
		assert.EqualError(t, err, `while extracting connection URL from body: while decoding json: invalid character 's' looking for beginning of value`)
	})
}

func TestApplicationService_Create(t *testing.T) {
	// GIVEN
	aCli := appFakeCli.NewSimpleClientset()
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}

	svc, err := application.NewApplicationService(application.Config{}, aCli.ApplicationconnectorV1alpha1(), nil, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)

	// WHEN
	app, err := svc.Create(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, app.Name, fixName)
	assert.Equal(t, app.Spec.Description, fixDesc)
	assert.Equal(t, app.Spec.Labels, fixLabels)
}

func TestApplicationService_Delete(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	aCli := appFakeCli.NewSimpleClientset(fixApplicationCR(fixName))

	svc, err := application.NewApplicationService(application.Config{}, aCli.ApplicationconnectorV1alpha1(), nil, newDummyInformer(), nil, newDummyInformer())
	require.NoError(t, err)

	// WHEN
	err = svc.Delete(fixName)

	// THEN
	require.NoError(t, err)
	_, err = aCli.ApplicationconnectorV1alpha1().Applications().Get(fixName, v1.GetOptions{})
	assert.True(t, apiErrors.IsNotFound(err))
}

func TestApplicationService_Update(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixApplicationCR(fixName))

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	svc, err := application.NewApplicationService(application.Config{}, aCli.ApplicationconnectorV1alpha1(), nil, newDummyInformer(), nil, aInformer)
	require.NoError(t, err)

	// WHEN
	app, err := svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixLabels, app.Spec.Labels)
	assert.Equal(t, fixDesc, app.Spec.Description)
}

func TestApplicationService_Update_ErrorInRetryLoop(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixApplicationCR(fixName))
	aCli.PrependReactor("update", "applications", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("fix")
	})

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	svc, err := application.NewApplicationService(application.Config{}, aCli.ApplicationconnectorV1alpha1(), nil, newDummyInformer(), nil, aInformer)
	require.NoError(t, err)

	// WHEN
	_, err = svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while updating %s [%s]: fix", pretty.Application, fixName))
}

func TestApplicationService_Update_SuccessAfterRetry(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	fixDesc := "desc"
	fixLabels := map[string]string{
		"fix": "lab",
	}

	// Application
	aCli := appFakeCli.NewSimpleClientset(fixApplicationCR(fixName))
	i := 0
	aCli.PrependReactor("update", "applications", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		if i < 3 {
			i++
			return true, nil, apiErrors.NewConflict(schema.GroupResource{}, "", errors.New("fix"))
		}
		return false, fixApplicationCR(fixName), nil
	})

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, 0)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, aInformer)

	svc, err := application.NewApplicationService(application.Config{}, aCli.ApplicationconnectorV1alpha1(), nil, newDummyInformer(), nil, aInformer)
	require.NoError(t, err)

	// WHEN
	app, err := svc.Update(fixName, fixDesc, fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixLabels, app.Spec.Labels)
	assert.Equal(t, fixDesc, app.Spec.Description)
}

func TestApplicationService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appListener := listener.NewApplication(nil, nil)
		svc.Subscribe(appListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appLister := listener.NewApplication(nil, nil)

		svc.Subscribe(appLister)
		svc.Subscribe(appLister)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appListerA := listener.NewApplication(nil, nil)
		appListerB := listener.NewApplication(nil, nil)

		svc.Subscribe(appListerA)
		svc.Subscribe(appListerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestApplicationService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appLister := listener.NewApplication(nil, nil)
		svc.Subscribe(appLister)

		svc.Unsubscribe(appLister)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appLister := listener.NewApplication(nil, nil)
		svc.Subscribe(appLister)
		svc.Subscribe(appLister)

		svc.Unsubscribe(appLister)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
		require.NoError(t, err)
		appListerA := listener.NewApplication(nil, nil)
		appListerB := listener.NewApplication(nil, nil)
		svc.Subscribe(appListerA)
		svc.Subscribe(appListerB)

		svc.Unsubscribe(appListerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := application.NewApplicationService(application.Config{}, nil, nil, newDummyInformer(), nil, newDummyInformer())
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

func fixApplicationMappingCR(name, ns string) mappingTypes.ApplicationMapping {
	return mappingTypes.ApplicationMapping{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

}

func fixApplicationCR(name string) *appTypes.Application {
	return &appTypes.Application{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}
