package k8sNew

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type LimitRangeList []*v1.LimitRange

func (r *Resolver) LimitRangesQuery(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	items := LimitRangeList{}

	//todo
	// err = r.Service().ListByIndex(apiRulesServiceAndHostnameIndex, apiRulesServiceAndHostnameIndexKey(namespace, serviceName, hostname), &items)

	return items, nil
}
