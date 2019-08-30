package docstopic

import "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"

func FindSource() func(slice []v1alpha1.Source, sourceName v1alpha1.DocsTopicSourceName, sourceType v1alpha1.DocsTopicSourceType) *v1alpha1.Source {
	return findSource
}
