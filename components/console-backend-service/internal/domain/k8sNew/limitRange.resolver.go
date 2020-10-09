package k8sNew

import (
	"context"
	"errors"
	"fmt"

	apimachinery "k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	v1 "k8s.io/api/core/v1"
)

type LimitRangeList []*v1.LimitRange

func (l *LimitRangeList) Append() interface{} {
	e := &v1.LimitRange{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) LimitRangesQuery(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	items := LimitRangeList{}
	err := r.LimitRangesService().ListInNamespace(namespace, &items)

	return items, err
}

func (r *Resolver) JsonField(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
	return resource.ToJson(obj)
}

func (r *Resolver) UpdateLimitRange(ctx context.Context, namespace string, name string, newJSON gqlschema.JSON) (*v1.LimitRange, error) {
	unstructured, unstructuredParseError := resource.ToUnstructured(&newJSON)
	if unstructuredParseError != nil {
		return nil, errors.New(fmt.Sprintf("could not parse input JSON to unstructured %s", unstructuredParseError))
	}

	newLimitRange := &v1.LimitRange{}
	jsonParseError := resource.FromUnstructured(unstructured, newLimitRange)
	if jsonParseError != nil {
		return nil, errors.New(fmt.Sprintf("could not convert LimitRange from unstructured %s", jsonParseError))
	}

	result := &v1.LimitRange{}
	err := r.LimitRangesService().Apply(newLimitRange, result)
	return result, err
}

type resourceLimitsItem interface {
	Memory() *apimachinery.Quantity
	Cpu() *apimachinery.Quantity
}

func (r *Resolver) GetResourceLimits(item resourceLimitsItem) (*gqlschema.ResourceLimits, error) {
	mem := item.Memory().String()
	cpu := item.Cpu().String()

	return &gqlschema.ResourceLimits{
		Memory: &mem,
		CPU:    &cpu,
	}, nil
}
