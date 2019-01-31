package gqlschema

import (
	"encoding/json"
	"io"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type JSON map[string]interface{}

func (y *JSON) UnmarshalGQL(v interface{}) error {
	if in, ok := v.(string); ok {
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(in), &jsonMap)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while unmarshalling %+v scalar object", y))
			return errors.Wrapf(err, "while unmarshalling %+v scalar object", y)
		}
		v = jsonMap
	}

	value := v.(map[string]interface{})
	*y = value
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (y JSON) MarshalGQL(w io.Writer) {
	bytes, err := json.Marshal(y)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while marshalling %+v scalar object", y))
	}
	w.Write(bytes)
}
