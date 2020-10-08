package k8sNew

import (
	"context"
	"encoding/json"
	"errors"

	apimachinery "k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	errs "github.com/pkg/errors"
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
	if obj == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(obj)
	if err != nil {
		return nil, errs.Wrapf(err, "while marshalling apirule `%s`", obj.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errs.Wrapf(err, "while unmarshalling apirule `%s` to map", obj.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errs.Wrapf(err, "while unmarshalling apirule `%s` to GQL JSON", obj.Name)
	}

	return result, nil
}

func (r *Resolver) UpdateLimitRange(ctx context.Context, name string, namespace string, generation int64, newJSON gqlschema.JSON) (*v1.LimitRange, error) {
	var newLimitRange v1.LimitRange

	unstructured, unstructuredParseError := resource.ToUnstructured(&newJSON)
	jsonParseError := resource.FromUnstructured(unstructured, &newLimitRange)
	if jsonParseError != nil || unstructuredParseError != nil {
		return nil, errors.New("could not parse input JSON")
	}

	result := &v1.LimitRange{}
	err := r.LimitRangesService().UpdateInNamespace(name, namespace, result, func() error {
		if result.Generation > generation {
			return errors.New("resource already modified")
		}

		result = &newLimitRange
		return nil
	})
	return result, err
}

type ResourceLimitsItem interface {
	Memory() *apimachinery.Quantity
	Cpu() *apimachinery.Quantity
}

func (r *Resolver) GetResourceLimits(item ResourceLimitsItem) (*gqlschema.ResourceLimits, error) {
	mem := item.Memory().String()
	cpu := item.Cpu().String()

	return &gqlschema.ResourceLimits{
		Memory: &mem,
		CPU:    &cpu,
	}, nil
}
