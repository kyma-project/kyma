package k8sNew

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type LimitRangeList []*v1.LimitRange

func (l *LimitRangeList) Append() interface{} {

	fmt.Println("adding")
	e := &v1.LimitRange{}
	return e
	*l = append(*l, e)
	return e
}

func (r *Resolver) LimitRangesQuery(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	items := LimitRangeList{}
	err := r.Service().ListInNamespace(namespace, &items)

	return items, err
}
