package docstopic

import "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"

func FindSource() func(slice []v1alpha1.Source, sourceName, sourceType string) *v1alpha1.Source {
	return findSource
}
