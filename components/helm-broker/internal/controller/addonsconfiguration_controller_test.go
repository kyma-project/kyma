package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	runtimeTypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileAddonsConfiguration_AddAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixAddonsConfiguration()
	ts := getTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			ts.bp.On("LoadCompleteBundle", e).
				Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{{Metadata: &chart.Metadata{Name: string(e.Name)}}}}, nil)

			ts.bundleStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &internal.Bundle{Name: internal.BundleName(e.Name)}).Return(false, nil)
			ts.chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).Return(false, nil)
		}
	}
	ts.bf.On("Exist", fixAddonsCfg.Namespace).Return(false, nil).Once()
	ts.bf.On("Create", fixAddonsCfg.Namespace).Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileAddonsConfiguration(ts.mgr, &ts.bp, &ts.bf, &ts.chartStorage, &ts.bundleStorage, true, &ts.dp, &ts.bs)

	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestReconcileAddonsConfiguration_UpdateAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixAddonsConfiguration()
	fixAddonsCfg.Generation = 1
	ts := getTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			ts.bp.On("LoadCompleteBundle", e).
				Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{{Metadata: &chart.Metadata{Name: string(e.Name)}}}}, nil)

			ts.bundleStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &internal.Bundle{Name: internal.BundleName(e.Name)}).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).
				Return(false, nil)
		}
	}
	ts.bf.On("Exist", fixAddonsCfg.Namespace).Return(false, nil).Once()
	ts.bf.On("Create", fixAddonsCfg.Namespace).Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileAddonsConfiguration(ts.mgr, &ts.bp, &ts.bf, &ts.chartStorage, &ts.bundleStorage, true, &ts.dp, &ts.bs)

	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestReconcileAddonsConfiguration_DeleteAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixDeletedAddonsConfiguration()
	ts := getTestSuite(t, fixAddonsCfg)

	ts.bf.On("Delete", fixAddonsCfg.Namespace).Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileAddonsConfiguration(ts.mgr, &ts.bp, &ts.bf, &ts.chartStorage, &ts.bundleStorage, true, &ts.dp, &ts.bs)

	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func fixAddonsConfiguration() *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				ReprocessRequest: 0,
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "http://example.com/index.yaml",
					},
				},
			},
		},
	}
}

func fixDeletedAddonsConfiguration() *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleted",
			Namespace:         "deleted",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
			Finalizers:        []string{v1alpha1.FinalizerAddonsConfiguration},
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				ReprocessRequest: 0,
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "http://example.com/index.yaml",
					},
				},
			},
		},
	}
}

func fixIndexDTO() *bundle.IndexDTO {
	return &bundle.IndexDTO{
		Entries: map[bundle.Name][]bundle.EntryDTO{
			"index": {
				{
					Name:        "redis",
					Version:     "0.1.0",
					Description: "desc",
				},
				{
					Name:        "testing",
					Version:     "0.1.0",
					Description: "desc",
				},
			},
		},
	}
}

type testSuite struct {
	t             *testing.T
	client        client.Client
	mgr           manager.Manager
	bp            automock.BundleProvider
	bf            automock.BrokerFacade
	dp            automock.DocsProvider
	bs            automock.BrokerSyncer
	bundleStorage automock.BundleStorage
	chartStorage  automock.ChartStorage
}

func getTestSuite(t *testing.T, objects ...runtime.Object) *testSuite {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, apis.AddToScheme(sch))
	require.NoError(t, v1beta1.AddToScheme(sch))
	require.NoError(t, v1.AddToScheme(sch))

	return &testSuite{
		t:   t,
		mgr: getFakeManager(t, fake.NewFakeClientWithScheme(sch, objects...), sch),
		bf:  automock.BrokerFacade{},
		bp:  automock.BundleProvider{},
		bs:  automock.BrokerSyncer{},
		dp:  automock.DocsProvider{},

		bundleStorage: automock.BundleStorage{},
		chartStorage:  automock.ChartStorage{},
	}
}

type fakeManager struct {
	t      *testing.T
	client client.Client
	sch    *runtime.Scheme
}

func (ts *testSuite) assertExpectations() {
	ts.bp.AssertExpectations(ts.t)
	ts.bf.AssertExpectations(ts.t)
	ts.dp.AssertExpectations(ts.t)
	ts.bs.AssertExpectations(ts.t)
	ts.bundleStorage.AssertExpectations(ts.t)
	ts.chartStorage.AssertExpectations(ts.t)
}

func (fakeManager) Add(manager.Runnable) error {
	return nil
}

func (fakeManager) SetFields(interface{}) error {
	return nil
}

func (fakeManager) Start(<-chan struct{}) error {
	return nil
}

func (fakeManager) GetConfig() *rest.Config {
	return &rest.Config{}
}

func (f *fakeManager) GetScheme() *runtime.Scheme {
	// Setup schemes for all resources
	return f.sch
}

func (fakeManager) GetAdmissionDecoder() runtimeTypes.Decoder {
	return nil
}

func (f *fakeManager) GetClient() client.Client {
	return f.client
}

func (fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (fakeManager) GetCache() cache.Cache {
	return nil
}

func (fakeManager) GetRecorder(name string) record.EventRecorder {
	return nil
}

func (fakeManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func getFakeManager(t *testing.T, cli client.Client, sch *runtime.Scheme) manager.Manager {
	return &fakeManager{
		t:      t,
		client: cli,
		sch:    sch,
	}
}
