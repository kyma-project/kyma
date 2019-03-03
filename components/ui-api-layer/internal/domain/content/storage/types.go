package storage

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

func (o *Content) UnmarshalJSON(jsonData []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return err
	}

	var data ContentData
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return err
	}

	o.Raw = raw
	o.Data = data

	return nil
}

type ApiSpec struct {
	Raw map[string]interface{}
}

func (o *ApiSpec) UnmarshalJSON(jsonData []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return err
	}

	o.Raw = raw

	return nil
}

type OpenApiSpec struct {
	Raw map[string]interface{}
}

func (o *OpenApiSpec) UnmarshalJSON(jsonData []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return err
	}

	o.Raw = raw

	return nil
}

type ODataSpec struct {
	Raw *string
}

func (o *ODataSpec) UnmarshalJSON(jsonData []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return err
	}

	if err != nil {
		str := string(jsonData)
		o.Raw = &str
	}

	return nil
}

func (o *ODataSpec) UnmarshalXML(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
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

func (o *AsyncApiSpec) UnmarshalJSON(jsonData []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return err
	}

	var data AsyncApiSpecData
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return err
	}

	o.Raw = raw
	o.Data = data

	return nil
}
