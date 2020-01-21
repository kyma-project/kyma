package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=deploymentGetter -output=automock -outpkg=automock -case=underscore
type deploymentGetter interface {
	Find(name string, namespace string) (*api.Deployment, error)
}

//go:generate mockery -name=limitRangeLister -output=automock -outpkg=automock -case=underscore
type limitRangeLister interface {
	List(ns string) ([]*v1.LimitRange, error)
	Create(namespace string, name string, limitRangeInput gqlschema.LimitRangeInput) (*v1.LimitRange, error)
}
