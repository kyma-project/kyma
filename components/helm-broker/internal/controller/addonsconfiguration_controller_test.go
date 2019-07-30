package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	runtimeTypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func TestReconcileAddonsConfiguration_AddAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixAddonsConfiguration()
	ts := getTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureDocsTopic", completeAddon.Addon, fixAddonsCfg.Namespace).Return(nil)
		}
	}
	ts.brokerFacade.On("Exist", fixAddonsCfg.Namespace).Return(false, nil).Once()
	ts.brokerFacade.On("Create", fixAddonsCfg.Namespace).Return(nil).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileAddonsConfiguration_AddAddonsProcess_ErrorIfBrokerExist(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixAddonsConfiguration()
	ts := getTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureDocsTopic", completeAddon.Addon, fixAddonsCfg.Namespace).Return(nil)
		}
	}
	ts.brokerFacade.On("Exist", fixAddonsCfg.Namespace).Return(false, errors.New("")).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.Error(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileAddonsConfiguration_UpdateAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixAddonsConfiguration()
	fixAddonsCfg.Generation = 2
	fixAddonsCfg.Status.ObservedGeneration = 1
	ts := getTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureDocsTopic", completeAddon.Addon, fixAddonsCfg.Namespace).Return(nil)
		}
	}
	ts.brokerFacade.On("Exist", fixAddonsCfg.Namespace).Return(false, nil).Once()
	ts.brokerFacade.On("Create", fixAddonsCfg.Namespace).Return(nil).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestReconcileAddonsConfiguration_UpdateAddonsProcess_ConflictingAddons(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixAddonsConfiguration()
	fixAddonsCfg.Generation = 2
	fixAddonsCfg.Status.ObservedGeneration = 1

	ts := getTestSuite(t, fixAddonsCfg, fixReadyAddonsConfiguration())
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")
			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)
		}
	}
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Status.ObservedGeneration, int64(2))
	assert.Equal(t, res.Status.Phase, v1alpha1.AddonsConfigurationFailed)
}

func TestReconcileAddonsConfiguration_DeleteAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixDeletedAddonsConfiguration()
	fixAddon := fixAddonWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Addon
	addonVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)
	ts := getTestSuite(t, fixAddonsCfg)

	ts.brokerFacade.On("Delete", fixAddonsCfg.Namespace).Return(nil).Once()
	ts.addonStorage.
		On("Get", internal.Namespace(fixAddonsCfg.Namespace), internal.AddonName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), addonVer).
		Return(fixAddon, nil)
	ts.addonStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Name, addonVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Name, fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Version).Return(nil)

	ts.docsProvider.On("EnsureDocsTopicRemoved", string(fixAddon.ID), fixAddonsCfg.Namespace).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.NotContains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileAddonsConfiguration_DeleteAddonsProcess_ReconcileOtherAddons(t *testing.T) {
	// GIVEN
	failedAddCfg := fixFailedAddonsConfiguration()
	fixAddonsCfg := fixDeletedAddonsConfiguration()
	fixAddon := fixAddonWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Addon
	addonVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)
	ts := getTestSuite(t, fixAddonsCfg, failedAddCfg)

	ts.brokerFacade.On("Delete", fixAddonsCfg.Namespace).Return(nil).Once()
	ts.addonStorage.
		On("Get", internal.Namespace(fixAddonsCfg.Namespace), internal.AddonName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), addonVer).
		Return(fixAddon, nil)
	ts.addonStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Name, addonVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Name, fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Version).Return(nil)

	ts.docsProvider.On("EnsureDocsTopicRemoved", string(fixAddon.ID), fixAddonsCfg.Namespace).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	otherAddon := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: failedAddCfg.Namespace, Name: failedAddCfg.Name}, &otherAddon)
	assert.NoError(t, err)
	assert.Equal(t, int(otherAddon.Spec.ReprocessRequest), 1)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.NotContains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileAddonsConfiguration_DeleteAddonsProcess_Error(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixDeletedAddonsConfiguration()
	ts := getTestSuite(t, fixAddonsCfg)

	ts.brokerFacade.On("Delete", fixAddonsCfg.Namespace).Return(errors.New("")).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.Error(t, err)
	assert.False(t, result.Requeue)
	assert.Equal(t, result.RequeueAfter, time.Second*15)

	res := v1alpha1.AddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
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

func fixFailedAddonsConfiguration() *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "failed",
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
		Status: v1alpha1.AddonsConfigurationStatus{
			CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
				Phase: v1alpha1.AddonsConfigurationFailed,
				Repositories: []v1alpha1.StatusRepository{
					{
						Status: v1alpha1.RepositoryStatusFailed,
						Addons: []v1alpha1.Addon{
							{
								Status:  v1alpha1.AddonStatusFailed,
								Name:    "redis",
								Version: "0.0.1",
							},
						},
					},
				},
			},
		},
	}
}

func fixReadyAddonsConfiguration() *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ready",
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
		Status: v1alpha1.AddonsConfigurationStatus{
			CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
				Phase: v1alpha1.AddonsConfigurationReady,
				Repositories: []v1alpha1.StatusRepository{
					{
						Status: v1alpha1.RepositoryStatusReady,
						Addons: []v1alpha1.Addon{
							{
								Status:  v1alpha1.AddonStatusReady,
								Name:    "redis",
								Version: "0.0.1",
							},
						},
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
			Namespace:         "test",
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
		Status: v1alpha1.AddonsConfigurationStatus{
			CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
				Phase: v1alpha1.AddonsConfigurationReady,
				Repositories: []v1alpha1.StatusRepository{
					{
						Status: v1alpha1.RepositoryStatusReady,
						Addons: []v1alpha1.Addon{
							{
								Status:  v1alpha1.AddonStatusReady,
								Name:    "redis",
								Version: "0.0.1",
							},
						},
					},
				},
			},
		},
	}
}

func fixIndexDTO() *addon.IndexDTO {
	return &addon.IndexDTO{
		Entries: map[addon.Name][]addon.EntryDTO{
			"redis": {
				{
					Name:        "redis",
					Version:     "0.0.1",
					Description: "desc",
				}},
			"testing": {
				{
					Name:        "testing",
					Version:     "0.0.1",
					Description: "desc",
				},
			},
		},
	}
}

type testSuite struct {
	t                  *testing.T
	mgr                manager.Manager
	addonGetterFactory *automock.AddonGetterFactory
	concreteGetter     *automock.AddonGetter
	brokerFacade       *automock.BrokerFacade
	docsProvider       *automock.DocsProvider
	brokerSyncer       *automock.BrokerSyncer
	addonStorage       *automock.AddonStorage
	chartStorage       *automock.ChartStorage
}

func getTestSuite(t *testing.T, objects ...runtime.Object) *testSuite {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, apis.AddToScheme(sch))
	require.NoError(t, v1beta1.AddToScheme(sch))
	require.NoError(t, v1.AddToScheme(sch))

	return &testSuite{
		t:                  t,
		mgr:                getFakeManager(t, fake.NewFakeClientWithScheme(sch, objects...), sch),
		brokerFacade:       &automock.BrokerFacade{},
		addonGetterFactory: &automock.AddonGetterFactory{},
		concreteGetter:     &automock.AddonGetter{},
		brokerSyncer:       &automock.BrokerSyncer{},
		docsProvider:       &automock.DocsProvider{},

		addonStorage: &automock.AddonStorage{},
		chartStorage: &automock.ChartStorage{},
	}
}

type fakeManager struct {
	t      *testing.T
	client client.Client
	sch    *runtime.Scheme
}

func (ts *testSuite) assertExpectations() {
	ts.concreteGetter.AssertExpectations(ts.t)
	ts.brokerFacade.AssertExpectations(ts.t)
	ts.docsProvider.AssertExpectations(ts.t)
	ts.brokerSyncer.AssertExpectations(ts.t)
	ts.addonStorage.AssertExpectations(ts.t)
	ts.chartStorage.AssertExpectations(ts.t)
	ts.addonGetterFactory.AssertExpectations(ts.t)
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
