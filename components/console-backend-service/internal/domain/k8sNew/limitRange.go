package k8sNew

import "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(limitRangesGroupVersionResource)
}
