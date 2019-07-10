package controller

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	runtimeTypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func TestReconcileAddonsConfiguration_ReconcileAddAddonsProcess(t *testing.T) {
	// Given
	//mgr := getFakeManager()
	//bp := mocks.BundleProvider{}
	//bf := mocks.BrokerFacade{}
	//s := mocks.Factory{}
	//dp := mocks.DocsProvider{}
	//bs := mocks.BrokerSyncer{}
	//
	//trac := NewReconcileAddonsConfiguration(mgr, &bp, &bf, &s, false, &dp, &bs)
	//
	//_, err := trac.Reconcile(reconcile.Request{types.NamespacedName{Namespace: "test", Name: "addon-test"}})
	//assert.NoError(t, err)
}

type fakeManager struct{}

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

func (fakeManager) GetScheme() *runtime.Scheme {
	return nil
}

func (fakeManager) GetAdmissionDecoder() runtimeTypes.Decoder {
	return nil
}

func (fakeManager) GetClient() client.Client {
	return fake.NewFakeClient(&v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "addon-test",
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
	})
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

func getFakeManager() manager.Manager {
	return &fakeManager{}
}
