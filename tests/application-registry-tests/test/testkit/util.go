package testkit

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	k8sResourceNameMaxLength = 64
	requestParamsNameFormat = "params-%s"
)

func GenerateIdentifier() string {
	return fmt.Sprintf("identifier-%s", rand.String(8))
}

func CreateParamsSecretName(resourceName string) string {
	paramsSecretName := fmt.Sprintf(requestParamsNameFormat, resourceName)
	if len(paramsSecretName) > k8sResourceNameMaxLength {
		paramsSecretName = paramsSecretName[0 : k8sResourceNameMaxLength-1]
	}

	return paramsSecretName
}
