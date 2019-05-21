package gqlschema

import (
	"encoding/json"
	"io"

	"github.com/golang/glog"

	"github.com/pkg/errors"
)

type ApplicationMappingService struct {
	ID string `json:"id"`
}

func (ams *ApplicationMappingService) UnmarshalGQL(input interface{}) error {
	if input == nil {
		return nil
	}
	value, ok := input.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected services type: %T, should be map[string]string", input)
	}

	idval, ok := value["id"]
	if !ok {
		return errors.New("missing services id parameter")
	}
	id, ok := idval.(string)
	if !ok {
		return errors.Errorf("unexpected services type: %T, should be string", idval)
	}

	ams.ID = id

	return nil
}

func (ams ApplicationMappingService) MarshalGQL(w io.Writer) {
	bytes, err := json.Marshal(ams)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while marshalling %+v scalar object", ams))
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while writing marshalled %+v object", ams))
		return
	}
}
