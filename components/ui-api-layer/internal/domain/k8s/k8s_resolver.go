package k8s

import "k8s.io/client-go/informers"

type k8sResolver struct {
	*environmentResolver
	*secretResolver
	*deploymentResolver
	*resourceQuotaResolver
	*resourceQuotaStatusResolver
	*limitRangeResolver

	informerFactory informers.SharedInformerFactory
}

func (r *k8sResolver) InformerFactory() informers.SharedInformerFactory {
	return r.informerFactory
}

