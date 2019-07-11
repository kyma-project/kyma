package controller

import (
	"testing"

	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileClusterAddonsConfiguration_AddAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixClusterAddonsConfiguration()
	ts := getClusterTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	ts.bundleStorage.On("FindAll", internal.ClusterWide).Return([]*internal.Bundle{}, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			ts.bp.On("LoadCompleteBundle", e).
				Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{{Metadata: &chart.Metadata{Name: string(e.Name)}}}}, nil)

			ts.bundleStorage.On("Upsert", internal.ClusterWide, &internal.Bundle{Name: internal.BundleName(e.Name)}).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).
				Return(false, nil)
		}
	}
	ts.bf.On("Exist").Return(false, nil).Once()
	ts.bf.On("Create").Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

	_, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
}

func TestReconcileClusterAddonsConfiguration_UpdateAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixClusterAddonsConfiguration()
	fixAddonsCfg.Generation = 1

	ts := getClusterTestSuite(t, fixAddonsCfg)
	indexDTO := fixIndexDTO()

	ts.bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	ts.bundleStorage.On("FindAll", internal.ClusterWide).Return([]*internal.Bundle{}, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			ts.bp.On("LoadCompleteBundle", e).
				Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{{Metadata: &chart.Metadata{Name: string(e.Name)}}}}, nil)

			ts.bundleStorage.On("Upsert", internal.ClusterWide, &internal.Bundle{Name: internal.BundleName(e.Name)}).
				Return(false, nil)
			ts.chartStorage.On("Upsert", internal.ClusterWide, &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).
				Return(false, nil)
		}
	}
	ts.bf.On("Exist").Return(false, nil).Once()
	ts.bf.On("Create").Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

	_, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
}

func TestReconcileClusterAddonsConfiguration_DeleteAddonsProcess(t *testing.T) {
	// Given
	fixAddonsCfg := fixDeletedClusterAddonsConfiguration()
	ts := getClusterTestSuite(t, fixAddonsCfg)

	ts.bf.On("Delete").Return(nil).Once()

	defer ts.assertExpectations()

	reconciler := NewReconcileClusterAddonsConfiguration(ts.mgr, &ts.bp, &ts.chartStorage, &ts.bundleStorage, &ts.bf, &ts.dp, &ts.bs, true)

	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
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
	}
}
