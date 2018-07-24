package kubeless

import "k8s.io/client-go/tools/cache"

func NewFunctionService(informer cache.SharedIndexInformer) *functionService {
	return newFunctionService(informer)
}

func NewFunctionResolver(functionSvc functionLister) *functionResolver {
	return newFunctionResolver(functionSvc)
}
