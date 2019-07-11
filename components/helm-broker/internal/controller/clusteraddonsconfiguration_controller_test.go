package controller

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileClusterAddonsConfiguration_ReconcileAddAddonsProcess(t *testing.T) {
	// Given
	mgr := getFakeManager(t)
	bp := automock.BundleProvider{}
	bf := automock.ClusterBrokerFacade{}
	dp := automock.ClusterDocsProvider{}
	bs := automock.ClusterBrokerSyncer{}
	bundleStorage := automock.BundleStorage{}
	chartStorage := automock.ChartStorage{}

	fixAddonsCfg := fixClusterAddonsConfiguration()
	indexDTO := fixIndexDTO()

	bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	bundleStorage.On("FindAll", internal.ClusterWide).Return([]*internal.Bundle{}, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			bp.On("LoadCompleteBundle", e).
				Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{{Metadata: &chart.Metadata{Name: string(e.Name)}}}}, nil)

			bundleStorage.On("Upsert", internal.ClusterWide, &internal.Bundle{Name: internal.BundleName(e.Name)}).Return(false, nil)
			chartStorage.On("Upsert", internal.ClusterWide, &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).Return(false, nil)
		}
	}
	bf.On("Exist").Return(false, nil).Once()
	bf.On("Create").Return(nil).Once()

	defer func() {
		bp.AssertExpectations(t)
		bf.AssertExpectations(t)
		dp.AssertExpectations(t)
		bs.AssertExpectations(t)
		bundleStorage.AssertExpectations(t)
		chartStorage.AssertExpectations(t)
	}()

	reconciler := NewReconcileClusterAddonsConfiguration(mgr, &bp, &chartStorage, &bundleStorage, &bf, &dp, &bs, true)

	_, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
}

func fixClusterAddonsConfiguration() *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
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
