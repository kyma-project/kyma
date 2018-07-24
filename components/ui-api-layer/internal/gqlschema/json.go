package gqlschema

import (
	"encoding/json"
	"io"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type JSON map[string]interface{}

func (y *JSON) UnmarshalGQL(v interface{}) error {
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
