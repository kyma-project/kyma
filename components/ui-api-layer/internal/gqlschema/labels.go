package gqlschema

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type Labels map[string]string

func (y *Labels) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}
	value, ok := v.(map[string]interface{})
	if !ok {
		return gqlerror.NewInternal(gqlerror.WithDetails(fmt.Sprintf("unexpected labels type: %T, should be map[string]string", v)))
	}

	labels, err := y.convertToLabels(value)
	if err != nil {
		return gqlerror.NewInternal(gqlerror.WithDetails(fmt.Sprintf("while converting labels: %v", err)))
	}
	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {
	bytes, err := json.Marshal(y)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while marshalling %+v scalar object", y))
		return
	}
	w.Write(bytes)
}

func (y *Labels) convertToLabels(labels map[string]interface{}) (Labels, error) {
	result := make(map[string]string)
	for k, v := range labels {
		val, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("given value `%v` must be a string", v)
		}
		result[k] = val
	}
	return result, nil
}
