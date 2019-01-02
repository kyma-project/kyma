package kubeless

import (
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"k8s.io/client-go/tools/cache"
)

func NewFunctionService(informer cache.SharedIndexInformer) *functionService {
	return newFunctionService(informer)
}

func NewFunctionResolver(functionSvc functionLister) (*functionResolver, error) {
	return newFunctionResolver(functionSvc)
}

func (r *PluggableResolver) SetFakeClient() {
	r.cfg.client = fake.NewSimpleClientset()
}