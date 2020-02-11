package k8s

import (
	"github.com/blang/semver"
	"strings"
)

type kymaVersionConverter struct{}


func (c *kymaVersionConverter) ToKymaVersion(deploymentImage string) string {
	deploymentImageSeparated := strings.FieldsFunc(deploymentImage, split)

	source := deploymentImageSeparated[0]
	if source != "eu.gcr.io" {
		return deploymentImage
	}

	version := deploymentImageSeparated[len(deploymentImageSeparated)-1]
	_, err := semver.Parse(version)
	if err != nil {
		branch := "master"
		if strings.HasPrefix(version,"PR-") {
			branch = "pull request"
		}
		return strings.Join([]string{branch, version}, " ")
	}

	return version
}

func split(r rune) bool {
	return r == '/' || r == ':'
}