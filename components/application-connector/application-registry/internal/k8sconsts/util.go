package k8sconsts

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func CreateOwnerReferenceForApplication(applicationName string, applicationUID types.UID) []v1.OwnerReference {
	return []v1.OwnerReference{
		{
			APIVersion: APIVersionApplication,
			Kind:       KindApplication,
			Name:       applicationName,
			UID:        applicationUID,
		},
	}
}
