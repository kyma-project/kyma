package k8s

import (
	"strings"

	"github.com/blang/semver"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/apps/v1"
)

type versionInfoConverter struct{}

func (c *versionInfoConverter) ToGQL(in *v1.Deployment) gqlschema.VersionInfo {
	deploymentImage := in.Spec.Template.Spec.Containers[0].Image
	deploymentImageSeparated := strings.FieldsFunc(deploymentImage, split)

	source := deploymentImageSeparated[0]
	if source != "eu.gcr.io" {
		return gqlschema.VersionInfo{
			KymaVersion: deploymentImage,
		}
	}

	version := deploymentImageSeparated[len(deploymentImageSeparated)-1]
	_, err := semver.Parse(version)
	if err != nil {
		branch := "master"
		if strings.HasPrefix(version, "PR-") {
			branch = "pull request"
		}
		return gqlschema.VersionInfo{
			KymaVersion: strings.Join([]string{branch, version}, " "),
		}
	}

	return gqlschema.VersionInfo{
		KymaVersion: version,
	}
}

func split(r rune) bool {
	return r == '/' || r == ':'
}
