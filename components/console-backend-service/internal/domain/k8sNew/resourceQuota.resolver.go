package k8sNew

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	v1 "k8s.io/api/core/v1"
	amResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *Resolver) GetHardField(item *v1.ResourceQuotaSpec) (*gqlschema.ResourceQuotaHard, error) {
	limitsMemory := item.Hard["limits.memory"]
	limitsMemoryString := limitsMemory.String()

	requestsMemory := item.Hard["requests.memory"]
	requestsMemoryString := requestsMemory.String()

	return &gqlschema.ResourceQuotaHard{
		Limits: &gqlschema.ResourceLimits{
			Memory: &limitsMemoryString,
		},
		Requests: &gqlschema.ResourceLimits{
			Memory: &requestsMemoryString,
		},
		Pods: item.Hard.Pods().String(),
	}, nil
}

func (r *Resolver) ResourceQuotaJSONfield(ctx context.Context, obj *v1.ResourceQuota) (gqlschema.JSON, error) {
	return resource.ToJson(obj)
}

func (r *Resolver) CreateResourceQuota(ctx context.Context, namespace string, name string, input gqlschema.ResourceQuotaInput) (*v1.ResourceQuota, error) {
	var errs apierror.ErrorFieldAggregate
	memoryLimitsParsed, err := amResource.ParseQuantity(*input.Limits.Memory)
	if err != nil {
		errs = append(errs, apierror.NewInvalidField("limits.memory", *input.Limits.Memory, fmt.Sprintf("while parsing %s memory limits", pretty.ResourceQuota)))
	}

	memoryRequestsParsed, err := amResource.ParseQuantity(*input.Requests.Memory)
	if err != nil {
		errs = append(errs, apierror.NewInvalidField("requests.memory", *input.Requests.Memory, fmt.Sprintf("while parsing %s memory requests", pretty.ResourceQuota)))
	}

	if len(errs) > 0 {
		return nil, apierror.NewInvalid(pretty.ResourceQuota, errs)
	}

	newResourceQuota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceLimitsMemory:   memoryLimitsParsed,
				v1.ResourceRequestsMemory: memoryRequestsParsed,
			},
		},
	}

	result := &v1.ResourceQuota{}
	creationError := r.ResourceQuotasService().Create(newResourceQuota, result)

	return result, creationError

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
