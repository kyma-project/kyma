package k8sNew

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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
