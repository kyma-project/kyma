package k8s

import (
	api "k8s.io/api/apps/v1"
)

//go:generate mockery -name=deploymentGetter -output=automock -outpkg=automock -case=underscore
type deploymentGetter interface {
	Find(name string, namespace string) (*api.Deployment, error)
}
