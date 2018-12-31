package kubeless

import "k8s.io/client-go/tools/cache"

func NewFunctionService(informer cache.SharedIndexInformer) *functionService {
	return newFunctionService(informer)
}

func NewFunctionResolver(functionSvc functionLister) (*functionResolver, error) {
	return newFunctionResolver(functionSvc)
}
