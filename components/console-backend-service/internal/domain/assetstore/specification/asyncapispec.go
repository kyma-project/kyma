package specification

import (
	"encoding/json"
)

type AsyncApiSpec struct {
	Raw  map[string]interface{}
	Data AsyncApiSpecData
}

type AsyncApiSpecData struct {
	AsyncAPI string
	Topics   map[string]interface{}
}

func (o *AsyncApiSpec) Decode(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	var specData AsyncApiSpecData
	err = json.Unmarshal(data, &specData)
	if err != nil {
		return err
	}

	o.Raw = raw
	o.Data = specData

	return nil
}
