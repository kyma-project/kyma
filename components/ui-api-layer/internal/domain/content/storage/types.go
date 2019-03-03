package storage

import (
	"encoding/json"
	"encoding/xml"
)

type Content struct {
	Raw  map[string]interface{}
	Data ContentData
}

type ContentData struct {
	Description string     `json:"description"`
	DisplayName string     `json:"displayName"`
	Docs        []Document `json:"docs"`
	ID          string     `json:"id"`
	Type        string     `json:"type"`
}

type Document struct {
	Order    string `json:"order"`
	Source   string `json:"source"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Internal bool   `json:"internal,omitempty"`
}

func (o *Content) Decode(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	var contentData ContentData
	err = json.Unmarshal(data, &contentData)
	if err != nil {
		return err
	}

	o.Raw = raw
	o.Data = contentData

	return nil
}

type ApiSpec struct {
	Raw map[string]interface{}
}

func (o *ApiSpec) Decode(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		if isInvalidBeginningCharacterError(err) {
			return nil
		}
		return err
	}

	o.Raw = raw

	return nil
}

type OpenApiSpec struct {
	Raw map[string]interface{}
}

func (o *OpenApiSpec) Decode(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		if isInvalidBeginningCharacterError(err) {
			return nil
		}
		return err
	}

	o.Raw = raw

	return nil
}

type ODataSpec struct {
	Raw *string
}

func (o *ODataSpec) Decode(data []byte) error {
	err := o.unmarshalJSON(data)
	if err == nil {
		return nil
	} else if err != nil && !isInvalidBeginningCharacterError(err) {
		return err
	}

	return o.unmarshalXML(data)
}

func (o *ODataSpec) unmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	str := string(data)
	o.Raw = &str

	return nil
}

func (o *ODataSpec) unmarshalXML(data []byte) error {
	var raw interface{}
	err := xml.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	str := string(data)
	o.Raw = &str

	return nil
}

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

func isInvalidBeginningCharacterError(err error) bool {
	switch err := err.(type) {
	case *json.SyntaxError:
		return err.Offset == 1 && err.Error() == "invalid character '<' looking for beginning of value"
	default:
		return false
	}
}
