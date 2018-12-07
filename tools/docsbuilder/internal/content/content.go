package content

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Content represents a single document topic
type Content struct {
	Name      string `yaml:"name"`
	Directory string `yaml:"directory"`
}

// Read reads documents from yaml file
func Read(path string) ([]Content, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading file %s", path)
	}

	var contentSlice []Content
	err = yaml.Unmarshal(yamlFile, &contentSlice)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling content of the file %s", path)
	}

	return contentSlice, nil
}

// ConstructPath constructs path for a single document topic
func ConstructPath(content Content, contentDirPath string) string {
	contentDir := content.Name
	if content.Directory != "" {
		contentDir = content.Directory
	}

	dir := fmt.Sprintf("%s/%s", contentDirPath, contentDir)
	return dir
}
