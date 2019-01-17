package k8s

import (
	api "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=deploymentGetter -output=automock -outpkg=automock -case=underscore
type deploymentGetter interface {
	Find(name string, environment string) (*api.Deployment, error)
}

//go:generate mockery -name=limitRangeLister -output=automock -outpkg=automock -case=underscore
type limitRangeLister interface {
	List(env string) ([]*v1.LimitRange, error)
}
