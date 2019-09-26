package spec

import (
	"encoding/json"
)

type AsyncAPISpec struct {
	Raw  map[string]interface{}
	Data AsyncAPISpecData
}

type AsyncAPISpecData struct {
	AsyncAPI string                 `json:"asyncapi"`
	Channels map[string]interface{} `json:"channels"`
}

func (o *AsyncAPISpec) Decode(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	var specData AsyncAPISpecData
	err = json.Unmarshal(data, &specData)
	if err != nil {
		return err
	}

	o.Raw = raw
	o.Data = specData

	return nil
}
