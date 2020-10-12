package k8sNew

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
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

	return &gqlschema.ResourceQuotaHard{
		Memory: &mem,
		CPU:    &cpu,
		Pods:   &pods,
	}, nil
}

func (r *Resolver) ResourceQuotaJSONfield(ctx context.Context, obj *v1.ResourceQuota) (gqlschema.JSON, error) {
	return resource.ToJson(obj)
}

func (r *Resolver) UpdateResourceQuota(ctx context.Context, namespace string, name string, newJSON gqlschema.JSON) (*v1.ResourceQuota, error) {
	unstructured, unstructuredParseError := resource.ToUnstructured(&newJSON)
	if unstructuredParseError != nil {
		return nil, errors.New(fmt.Sprintf("could not parse input JSON to unstructured %s", unstructuredParseError))
	}

	newResourceQuota := &v1.ResourceQuota{}
	jsonParseError := resource.FromUnstructured(unstructured, newResourceQuota)
	if jsonParseError != nil {
		return nil, errors.New(fmt.Sprintf("could not convert ResourceQuota from unstructured %s", jsonParseError))
	}

	result := &v1.ResourceQuota{}
	err := r.ResourceQuotasService().Apply(newResourceQuota, result)
	return result, err
}
