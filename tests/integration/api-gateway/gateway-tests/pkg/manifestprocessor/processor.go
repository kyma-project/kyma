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

//ParseFromFile parse simple yaml manifests
func ParseFromFile(fileName string, directory string, separator string) ([]unstructured.Unstructured, error) {
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

//ParseFromFileWithTemplate parse manifests with goTemplate support
func ParseFromFileWithTemplate(fileName string, directory string, separator string, templateData interface{}) ([]unstructured.Unstructured, error) {
	manifestsRaw, err := convert(getManifestsFromFile(fileName, directory, separator))
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, raw := range manifestsRaw {
		man := parseTemplateWithData(raw, templateData)
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}
