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

func TestReconcileAddonsConfiguration_ReconcileAddAddonsProcess(t *testing.T) {
	// Given
	mgr := getFakeManager(t)
	bp := automock.BundleProvider{}
	bf := automock.BrokerFacade{}
	dp := automock.DocsProvider{}
	bs := automock.BrokerSyncer{}
	bundleStorage := automock.BundleStorage{}
	chartStorage := automock.ChartStorage{}

	fixAddonsCfg := fixAddonsConfiguration()
	indexDTO := fixIndexDTO()

	bp.On("GetIndex", fixAddonsCfg.Spec.Repositories[0].URL).Return(indexDTO, nil)
	bundleStorage.On("FindAll", internal.Namespace(fixAddonsCfg.Namespace)).Return([]*internal.Bundle{}, nil)

	for _, entry := range indexDTO.Entries {
		for _, e := range entry {
			bp.On("LoadCompleteBundle", e).Return(bundle.CompleteBundle{Bundle: &internal.Bundle{Name: internal.BundleName(e.Name)}, Charts: []*chart.Chart{
				{
					Metadata: &chart.Metadata{
						Name: string(e.Name),
					},
				},
			}}, nil)
			bundleStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &internal.Bundle{Name: internal.BundleName(e.Name)}).Return(false, nil)
			chartStorage.On("Upsert", internal.Namespace(fixAddonsCfg.Namespace), &chart.Chart{Metadata: &chart.Metadata{Name: string(e.Name)}}).Return(false, nil)
		}
	}
	bf.On("Exist", fixAddonsCfg.Namespace).Return(false, nil).Once()
	bf.On("Create", fixAddonsCfg.Namespace).Return(nil).Once()

	defer func() {
		bp.AssertExpectations(t)
		bf.AssertExpectations(t)
		dp.AssertExpectations(t)
		bs.AssertExpectations(t)
		bundleStorage.AssertExpectations(t)
		chartStorage.AssertExpectations(t)
	}()

	reconciler := NewReconcileAddonsConfiguration(mgr, &bp, &bf, &chartStorage, &bundleStorage, true, &dp, &bs)

	_, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
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

type fakeManager struct {
	t *testing.T
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
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(f.t, err)
	require.NoError(f.t, apis.AddToScheme(sch))
	require.NoError(f.t, v1beta1.AddToScheme(sch))
	require.NoError(f.t, v1.AddToScheme(sch))
	return sch
}

func (fakeManager) GetAdmissionDecoder() runtimeTypes.Decoder {
	return nil
}

func (f *fakeManager) GetClient() client.Client {
	return fake.NewFakeClientWithScheme(f.GetScheme(), fixAddonsConfiguration(), fixClusterAddonsConfiguration())
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

func getFakeManager(t *testing.T) manager.Manager {
	return &fakeManager{t: t}
}
