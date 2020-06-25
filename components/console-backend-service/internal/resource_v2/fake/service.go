package fake

import (
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicFake "k8s.io/client-go/dynamic/fake"
)

func NewSimpleFakeServiceFactory(informerResyncPeriod time.Duration) *resource_v2.ServiceFactory {
	client := dynamicFake.NewSimpleDynamicClient(runtime.NewScheme())
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(client, informerResyncPeriod)
	return resource_v2.NewServiceFactory(client, informerFactory)
}

func NewFakeServiceFactory(addToScheme func(*runtime.Scheme) error, objects ...runtime.Object) (*resource_v2.ServiceFactory, error) {
	scheme := runtime.NewScheme()
	err := addToScheme(scheme)
	if err != nil {
		return nil, err
	}
	result := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		result[i], err = resource_v2.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}
	}
	client := dynamicFake.NewSimpleDynamicClient(scheme, result...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(client, 10)
	return resource_v2.NewServiceFactory(client, informerFactory), nil
}
