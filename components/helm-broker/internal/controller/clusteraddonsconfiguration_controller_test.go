package controller

import (
	"testing"

	"time"

	"context"
	"errors"

	"fmt"

	"github.com/Masterminds/semver"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
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

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeBundle := fixBundleWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.bp.On("LoadCompleteBundle", e).
				Return(completeBundle, nil)

			ts.bundleStorage.On("Upsert", internal.ClusterWide, completeBundle.Bundle).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeBundle.Charts[0]).
				Return(false, nil)
			ts.dp.On("EnsureClusterDocsTopic", completeBundle.Bundle).Return(nil)
		}
	}
	ts.bf.On("Exist").Return(false, nil).Once()
	ts.bf.On("Create").Return(nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeBundle := fixBundleWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.bp.On("LoadCompleteBundle", e).
				Return(completeBundle, nil)

			ts.bundleStorage.On("Upsert", internal.ClusterWide, completeBundle.Bundle).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeBundle.Charts[0]).
				Return(false, nil)
			ts.dp.On("EnsureClusterDocsTopic", completeBundle.Bundle).Return(nil)
		}
	}
	ts.bf.On("Exist").Return(false, errors.New("")).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeBundle := fixBundleWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")

			ts.bp.On("LoadCompleteBundle", e).Return(completeBundle, nil)

			ts.bundleStorage.On("Upsert", internal.ClusterWide, completeBundle.Bundle).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, completeBundle.Charts[0]).
				Return(false, nil)
			ts.dp.On("EnsureClusterDocsTopic", completeBundle.Bundle).Return(nil)
		}
	}
	ts.bf.On("Exist").Return(false, nil).Once()
	ts.bf.On("Create").Return(nil).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

	// THEN
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.False(t, result.Requeue)
	assert.NoError(t, err)

	res := v1alpha1.ClusterAddonsConfiguration{}
	err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Name: fixAddonsCfg.Name}, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Status.ObservedGeneration, int64(2))
}

func TestReconcileClusterAddonsConfiguration_UpdateAddonsProcess_ConflictingBundles(t *testing.T) {
	// GIVEN
	fixAddonsCfg := fixClusterAddonsConfiguration()
	fixAddonsCfg.Generation = 2
	fixAddonsCfg.Status.ObservedGeneration = 1

	ts := getClusterTestSuite(t, fixAddonsCfg, fixReadyClusterAddonsConfiguration())
	indexDTO := fixIndexDTO()
	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			completeBundle := fixBundleWithDocsURL(string(e.Name), string(e.Name), "example.com", "example.com")
			ts.bp.On("LoadCompleteBundle", e).Return(completeBundle, nil)
		}
	}
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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
	fixBundle := fixBundleWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Bundle
	bundleVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)

	ts := getClusterTestSuite(t, fixAddonsCfg)

	ts.bf.On("Delete").Return(nil).Once()
	ts.bundleStorage.
		On("Get", internal.ClusterWide, internal.BundleName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), bundleVer).
		Return(fixBundle, nil)
	ts.bundleStorage.On("Remove", internal.ClusterWide, fixBundle.Name, bundleVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixBundle.Plans[internal.BundlePlanID(fmt.Sprintf("plan-%s", fixBundle.Name))].ChartRef.Name, fixBundle.Plans[internal.BundlePlanID(fmt.Sprintf("plan-%s", fixBundle.Name))].ChartRef.Version).Return(nil)

	ts.dp.On("EnsureClusterDocsTopicRemoved", string(fixBundle.ID)).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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
	fixBundle := fixBundleWithEmptyDocs("id", fixAddonsCfg.Status.Repositories[0].Addons[0].Name, "example.com").Bundle
	bundleVer := *semver.MustParse(fixAddonsCfg.Status.Repositories[0].Addons[0].Version)

	ts := getClusterTestSuite(t, fixAddonsCfg, failedAddCfg)

	ts.bf.On("Delete").Return(nil).Once()
	ts.bundleStorage.
		On("Get", internal.ClusterWide, internal.BundleName(fixAddonsCfg.Status.Repositories[0].Addons[0].Name), bundleVer).
		Return(fixBundle, nil)
	ts.bundleStorage.On("Remove", internal.ClusterWide, fixBundle.Name, bundleVer).Return(nil)
	ts.chartStorage.On("Remove", internal.Namespace(fixAddonsCfg.Namespace), fixBundle.Plans[internal.BundlePlanID(fmt.Sprintf("plan-%s", fixBundle.Name))].ChartRef.Name, fixBundle.Plans[internal.BundlePlanID(fmt.Sprintf("plan-%s", fixBundle.Name))].ChartRef.Version).Return(nil)

	ts.dp.On("EnsureClusterDocsTopicRemoved", string(fixBundle.ID)).Return(nil)
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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

	ts.bf.On("Delete").Return(errors.New("")).Once()
	defer ts.assertExpectations()

	// WHEN
	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

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
	t             *testing.T
	mgr           manager.Manager
	bp            automock.BundleProvider
	bf            automock.ClusterBrokerFacade
	dp            automock.ClusterDocsProvider
	bs            automock.ClusterBrokerSyncer
	bundleStorage automock.BundleStorage
	chartStorage  automock.ChartStorage
}

func getClusterTestSuite(t *testing.T, objects ...runtime.Object) *clusterTestSuite {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, apis.AddToScheme(sch))
	require.NoError(t, v1beta1.AddToScheme(sch))
	require.NoError(t, v1.AddToScheme(sch))

	return &clusterTestSuite{
		t:   t,
		mgr: getFakeManager(t, fake.NewFakeClientWithScheme(sch, objects...), sch),
		bf:  automock.ClusterBrokerFacade{},
		bp:  automock.BundleProvider{},
		bs:  automock.ClusterBrokerSyncer{},
		dp:  automock.ClusterDocsProvider{},

		bundleStorage: automock.BundleStorage{},
		chartStorage:  automock.ChartStorage{},
	}
}

func (ts *clusterTestSuite) assertExpectations() {
	ts.dp.AssertExpectations(ts.t)
	ts.bf.AssertExpectations(ts.t)
	ts.bp.AssertExpectations(ts.t)
	ts.bs.AssertExpectations(ts.t)
	ts.bundleStorage.AssertExpectations(ts.t)
	ts.chartStorage.AssertExpectations(ts.t)
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
