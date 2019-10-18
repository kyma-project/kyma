package manifestprocessor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	separatorYAML = "---"
)

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
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

func getManifestsFromFile(fileName string, directory string, separator string) []string {
	if separator == "" {
		separator = separatorYAML
	}
	data, err := ioutil.ReadFile(directory + fileName)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), separator)
}

func convert(inputYAML []string) ([]string, error) {
	var result []string
	for _, input := range inputYAML {
		json, err := yaml.YAMLToJSON([]byte(input))
		if err != nil {
			return nil, err
		}
		if string(json) != "null" {
			result = append(result, string(json))
		}
	}
	return result, nil
}

func parseTemplateWithData(templateRaw string, data interface{}) string {
	tmpl, err := template.New("tmpl").Parse(templateRaw)
	if err != nil {
		panic(err)
	}
	var resource bytes.Buffer
	err = tmpl.Execute(&resource, data)
	if err != nil {
		panic(err)
	}
	return resource.String()
}

//Parse Parse simple yaml manifests
func Parse(fileName string, directory string, separator string) ([]unstructured.Unstructured, error) {
	manifests, err := convert(getManifestsFromFile(fileName, directory, separator))
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, man := range manifests {
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}

//ParseTemplate Parse manifests with goTemplate support
func ParseTemplate(fileName string, directory string, separator string, testID string) ([]unstructured.Unstructured, error) {
	manifestsRaw, err := convert(getManifestsFromFile(fileName, directory, separator))
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, raw := range manifestsRaw {
		man := parseTemplateWithData(raw, struct{ TestID string }{TestID: testID})
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}
