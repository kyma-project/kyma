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
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileClusterAddonsConfiguration_AddAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixClusterAddonsConfiguration()
	ts := getClusterTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.ClusterWide, completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureClusterDocsTopic", completeAddon.Addon).Return(nil)

		}
	}
	ts.brokerFacade.On("Exist").Return(false, nil).Once()
	ts.brokerFacade.On("Create").Return(nil).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "cluster-addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileClusterAddonsConfiguration_AddAddonsProcess_Error(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixClusterAddonsConfiguration()
	ts := getClusterTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).
				Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.ClusterWide, completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureClusterDocsTopic", completeAddon.Addon).Return(nil)
		}
	}
	ts.brokerFacade.On("Exist").Return(false, errors.New("")).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "cluster-addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.Error(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileClusterAddonsConfiguration_UpdateAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixClusterAddonsConfiguration()
	fixAddonsCfg.Generation = 2
	fixAddonsCfg.Status.ObservedGeneration = 1

	ts := getClusterTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()
	tmpDir := os.TempDir()

	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.concreteGetter.On("GetCompleteAddon", e).Return(completeAddon, nil)

			ts.addonStorage.On("Upsert", internal.ClusterWide, completeAddon.Addon).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeAddon.Charts[0]).
				Return(false, nil)
			ts.docsProvider.On("EnsureClusterDocsTopic", completeAddon.Addon).Return(nil)
		}

	}
	ts.brokerFacade.On("Exist").Return(false, nil).Once()
	ts.brokerFacade.On("Create").Return(nil).Once()
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "cluster-addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.False(t, result.Requeue)
	assert.NoError(t, err)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Status.ObservedGeneration, int64(2))
}

func TestReconcileClusterAddonsConfiguration_UpdateAddonsProcess_ConflictingAddons(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixClusterAddonsConfiguration()
	fixAddonsCfg.Generation = 2
	fixAddonsCfg.Status.ObservedGeneration = 1
	tmpDir := os.TempDir()

	ts := getClusterTestSuite(t, fixAddonsCfg, fixReadyClusterAddonsConfiguration())
	indexDTO := fixIndexDTO()
	ts.concreteGetter.On("GetIndex").Return(indexDTO, nil)
	ts.concreteGetter.On("Cleanup").Return(nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeAddon := fixAddonWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")
			ts.concreteGetter.On("GetCompleteAddon", e).Return(completeAddon, nil)
		}
	}
	ts.addonGetterFactory.On("NewGetter", fixAddonsCfg.Spec.Repositories[0].URL, path.Join(tmpDir, "cluster-addon-loader-dst")).Return(ts.concreteGetter, nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, tmpDir, spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Status.ObservedGeneration, int64(2))
	assert.Equal(t, res.Status.Phase, v1alpha1.AddonsConfigurationFailed)
}

func TestReconcileClusterAddonsConfiguration_DeleteAddonsProcess(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixDeletedClusterAddonsConfiguration()
	fixAddon := fixAddonWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Addon
	addonVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)

	ts := getClusterTestSuite(t, fixAddonsCfg)

	ts.brokerFacade.On("Delete").Return(nil).Once()
	ts.addonStorage.
		On("Get", internal.ClusterWide, internal.AddonName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), addonVer).
		Return(fixAddon, nil)
	ts.addonStorage.On("Remove", internal.ClusterWide, fixAddon.Name, addonVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Name, fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Version).Return(nil)

	ts.docsProvider.On("EnsureClusterDocsTopicRemoved", string(fixAddon.ID)).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.NotContains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileClusterAddonsConfiguration_DeleteAddonsProcess_ReconcileOtherAddons(t *testing.T) {
	// GIVEN
	failedAddCfg := fixFailedClusterAddonsConfiguration()
	fixAddonsCfg := fixDeletedClusterAddonsConfiguration()
	fixAddon := fixAddonWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Addon
	addonVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)

	ts := getClusterTestSuite(t, fixAddonsCfg, failedAddCfg)

	ts.brokerFacade.On("Delete").Return(nil).Once()
	ts.addonStorage.
		On("Get", internal.ClusterWide, internal.AddonName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), addonVer).
		Return(fixAddon, nil)
	ts.addonStorage.On("Remove", internal.ClusterWide, fixAddon.Name, addonVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Name, fixAddon.Plans[internal.AddonPlanID(fmt.Sprintf("plan-%s", fixAddon.Name))].ChartRef.Version).Return(nil)

	ts.docsProvider.On("EnsureClusterDocsTopicRemoved", string(fixAddon.ID)).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	otherAddon := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: failedAddCfg.Name}, &otherAddon)
	assert.NoError(t, err)
	assert.Equal(t, int(otherAddon.Spec.ReprocessRequest), 1)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.NotContains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

func TestReconcileClusterAddonsConfiguration_DeleteAddonsProcess_Error(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixDeletedClusterAddonsConfiguration()
	ts := getClusterTestSuite(t, fixAddonsCfg)

	ts.brokerFacade.On("Delete").Return(errors.New("")).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, ts.addonGetterFactory, ts.chartStorage, ts.addonStorage, ts.brokerFacade, ts.docsProvider, ts.brokerSyncer, os.TempDir(), spy.NewLogDummy())

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.Error(t, err)
	assert.False(t, result.Requeue)
	assert.Equal(t, result.RequeueAfter, time.Second*15)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Contains(t, res.Finalizers, v1alpha1.FinalizerAddonsConfiguration)
}

type clusterTestSuite struct {
	t                  *testing.T
	mgr                manager.Manager
	addonGetterFactory *automock.AddonGetterFactory
	concreteGetter     *automock.AddonGetter
	brokerFacade       *automock.ClusterBrokerFacade
	docsProvider       *automock.ClusterDocsProvider
	brokerSyncer       *automock.ClusterBrokerSyncer
	addonStorage       *automock.AddonStorage
	chartStorage       *automock.ChartStorage
}

func getClusterTestSuite(t *testing.T, objects ...runtime.Object) *clusterTestSuite {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, apis.AddToScheme(sch))
	require.NoError(t, v1beta1.AddToScheme(sch))
	require.NoError(t, v1.AddToScheme(sch))

	return &clusterTestSuite{
		t:                  t,
		mgr:                getFakeManager(t, fake.NewFakeClientWithScheme(sch, objects...), sch),
		brokerFacade:       &automock.ClusterBrokerFacade{},
		addonGetterFactory: &automock.AddonGetterFactory{},
		concreteGetter:     &automock.AddonGetter{},
		brokerSyncer:       &automock.ClusterBrokerSyncer{},
		docsProvider:       &automock.ClusterDocsProvider{},

		addonStorage: &automock.AddonStorage{},
		chartStorage: &automock.ChartStorage{},
	}
}

func (ts *clusterTestSuite) assertExpectations() {
	ts.docsProvider.AssertExpectations(ts.t)
	ts.brokerFacade.AssertExpectations(ts.t)
	ts.concreteGetter.AssertExpectations(ts.t)
	ts.brokerSyncer.AssertExpectations(ts.t)
	ts.addonStorage.AssertExpectations(ts.t)
	ts.chartStorage.AssertExpectations(ts.t)
	ts.addonGetterFactory.AssertExpectations(ts.t)
}

func fixClusterAddonsConfiguration() *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
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

func fixFailedClusterAddonsConfiguration() *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "failed",
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				ReprocessRequest: 0,
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "http://example.com/index.yaml",
					},
				},
			},
		},
		Status: v1alpha1.ClusterAddonsConfigurationStatus{
			CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
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

func fixReadyClusterAddonsConfiguration() *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ready",
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				ReprocessRequest: 0,
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "http://example.com/index.yaml",
					},
				},
			},
		},
		Status: v1alpha1.ClusterAddonsConfigurationStatus{
			CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
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

func fixDeletedClusterAddonsConfiguration() *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleted",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
			Finalizers:        []string{v1alpha1.FinalizerAddonsConfiguration},
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				ReprocessRequest: 0,
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "http://example.com/index.yaml",
					},
				},
			},
		},
		Status: v1alpha1.ClusterAddonsConfigurationStatus{
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
