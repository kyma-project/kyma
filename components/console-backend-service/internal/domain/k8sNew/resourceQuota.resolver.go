package k8sNew

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type ResourceQuotaList []*v1.ResourceQuota

func (l *ResourceQuotaList) Append() interface{} {
	e := &v1.ResourceQuota{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) ResourceQuotasQuery(ctx context.Context, namespace string) ([]*v1.ResourceQuota, error) {
	items := ResourceQuotaList{}
	err := r.ResourceQuotasService().ListInNamespace(namespace, &items)

	return items, err
}

func (r *Resolver) GetHardField(item v1.ResourceList) (*gqlschema.ResourceQuotaHard, error) {
	mem := item.Memory().String()
	cpu := item.Cpu().String()
	pods := item.Pods().String()

	// limits := gqlschema.ResourceLimits{
	// 	Memory: &mem,
	// 	CPU:    &cpu,
	// }

	return &gqlschema.ResourceQuotaHard{
		Memory: &mem,
		CPU:    &cpu,
		Pods:   &pods,
	}, nil
}

// func (r *Resolver) JsonField(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
// 	return resource.ToJson(obj)
// }

// func (r *Resolver) UpdateLimitRange(ctx context.Context, namespace string, name string, newJSON gqlschema.JSON) (*v1.LimitRange, error) {
// 	unstructured, unstructuredParseError := resource.ToUnstructured(&newJSON)
// 	if unstructuredParseError != nil {
// 		return nil, errors.New(fmt.Sprintf("could not parse input JSON to unstructured %s", unstructuredParseError))
// 	}

// 	newLimitRange := &v1.LimitRange{}
// 	jsonParseError := resource.FromUnstructured(unstructured, newLimitRange)
// 	if jsonParseError != nil {
// 		return nil, errors.New(fmt.Sprintf("could not convert LimitRange from unstructured %s", jsonParseError))
// 	}

// 	result := &v1.LimitRange{}
// 	err := r.LimitRangesService().Apply(newLimitRange, result)
// 	return result, err
// }

// type resourceLimitsItem interface {
// 	Memory() *apimachinery.Quantity
// 	Cpu() *apimachinery.Quantity
// }

// func (r *Resolver) GetResourceLimits(item resourceLimitsItem) (*gqlschema.ResourceLimits, error) {
// 	mem := item.Memory().String()
// 	cpu := item.Cpu().String()

// 	return &gqlschema.ResourceLimits{
// 		Memory: &mem,
// 		CPU:    &cpu,
// 	}, nil
// }
