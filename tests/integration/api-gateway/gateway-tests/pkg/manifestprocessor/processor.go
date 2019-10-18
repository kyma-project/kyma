package manifestprocessor

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	separatorYAML = "---"
)

// ParseManifest .
func ParseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

// GetManifestsFromFile .
func GetManifestsFromFile(fileName string, directory string, separator string) []string {
	if separator == "" {
		separator = separatorYAML
	}
	data, err := ioutil.ReadFile(directory + fileName)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), separator)
}
