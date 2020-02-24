package manifestprocessor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
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

func getManifestsFromFile(t *testing.T, fileName string, directory string, separator string) []string {
	data, err := ioutil.ReadFile(path.Join(directory, fileName))
	if err != nil {
		require.NoError(t, err)
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

func parseTemplateWithData(t *testing.T, templateRaw string, data interface{}) string {
	tmpl, err := template.New("tmpl").Parse(templateRaw)
	if err != nil {
		require.NoError(t, err)
	}
	var resource bytes.Buffer
	err = tmpl.Execute(&resource, data)
	if err != nil {
		require.NoError(t, err)
	}
	return resource.String()
}

//ParseFromFile parse simple yaml manifests
func ParseFromFile(t *testing.T, fileName string, directory string, separator string) ([]unstructured.Unstructured, error) {
	manifests, err := convert(getManifestsFromFile(t, fileName, directory, separator))
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
func ParseFromFileWithTemplate(t *testing.T, fileName string, directory string, separator string, templateData interface{}) ([]unstructured.Unstructured, error) {
	manifestsRaw, err := convert(getManifestsFromFile(t, fileName, directory, separator))
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, raw := range manifestsRaw {
		man := parseTemplateWithData(t, raw, templateData)
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}
