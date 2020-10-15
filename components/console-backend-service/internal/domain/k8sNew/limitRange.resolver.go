package k8sNew

import (
	"context"
	"errors"
	"fmt"

	amResource "k8s.io/apimachinery/pkg/api/resource"
	apimachinery "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
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

func (r *Resolver) LimitRangeJSONfield(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
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

func (r *Resolver) CreateLimitRange(ctx context.Context, namespace string, name string, input gqlschema.LimitRangeInput) (*v1.LimitRange, error) {
	var errs apierror.ErrorFieldAggregate
	defaultParsed, err := amResource.ParseQuantity(*input.Default.Memory)
	if err != nil {
		errs = append(errs, apierror.NewInvalidField("limits.memory", *input.Default.Memory, fmt.Sprintf("while parsing %s memory limits", pretty.LimitRange)))
	}

	defaultRequestParsed, err := amResource.ParseQuantity(*input.DefaultRequest.Memory)
	if err != nil {
		errs = append(errs, apierror.NewInvalidField("requests.memory", *input.DefaultRequest.Memory, fmt.Sprintf("while parsing %s memory requests", pretty.LimitRange)))
	}

	maxParsed, err := amResource.ParseQuantity(*input.Max.Memory)
	if err != nil {
		errs = append(errs, apierror.NewInvalidField("requests.memory", *input.Max.Memory, fmt.Sprintf("while parsing %s memory requests", pretty.LimitRange)))
	}

	if len(errs) > 0 {
		return nil, apierror.NewInvalid(pretty.LimitRange, errs)
	}

	newLimitRange := &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitType(input.Type),
					Default: v1.ResourceList{
						v1.ResourceMemory: defaultParsed,
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: defaultRequestParsed,
					},
					Max: v1.ResourceList{
						v1.ResourceMemory: maxParsed,
					},
				},
			},
		},
	}

	result := &v1.LimitRange{}
	creationError := r.LimitRangesService().Create(newLimitRange, result)

	return result, creationError

}

type resourceLimitsItem interface {
	Memory() *apimachinery.Quantity
	Cpu() *apimachinery.Quantity
}

func (r *Resolver) GetLimitRangeResources(item resourceLimitsItem) (*gqlschema.ResourceLimits, error) {
	mem := item.Memory().String()
	cpu := item.Cpu().String()

	return &gqlschema.ResourceLimits{
		Memory: &mem,
		CPU:    &cpu,
	}, nil
}
