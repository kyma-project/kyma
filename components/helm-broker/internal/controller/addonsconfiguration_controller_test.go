package controller

import (
	"testing"

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
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"

	"k8s.io/apimachinery/pkg/types"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"k8s.io/client-go/kubernetes/scheme"
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

	defer func() {
		bp.AssertExpectations(t)
		bf.AssertExpectations(t)
		dp.AssertExpectations(t)
		bs.AssertExpectations(t)
		bundleStorage.AssertExpectations(t)
		chartStorage.AssertExpectations(t)
	}()

	fixAddonsCfg := fixAddonsConfiguration()

	reconciler := NewReconcileAddonsConfiguration(mgr, &bp, &bf, &chartStorage, &bundleStorage, true, &dp, &bs)

	err := mgr.GetClient().Create(context.Background(), fixAddonsCfg)
	assert.NoError(t, err)

	_, err = reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Namespace: fixAddonsCfg.Namespace, Name: fixAddonsCfg.Name}})
	assert.NoError(t, err)
}

func fixAddonsConfiguration() *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
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
					{
						URL: "http://example.com/index-with-wrong-addons.yaml",
					},
					{
						URL: "http://example.com/wrong-index.yaml",
					},
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
	// Setup Scheme for all resources
	sch := scheme.Scheme
	assert.NoError(f.t, apis.AddToScheme(sch))
	assert.NoError(f.t, v1beta1.AddToScheme(sch))
	assert.NoError(f.t, v1alpha1.AddToScheme(sch))
	return sch
}

func (fakeManager) GetAdmissionDecoder() runtimeTypes.Decoder {
	return nil
}

func (fakeManager) GetClient() client.Client {
	return fake.NewFakeClient()
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
