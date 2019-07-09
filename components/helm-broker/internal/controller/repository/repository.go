package repository

import (
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

// RepositoryController is a wraper for StatusRepository
type RepositoryController struct {
	Repository addonsv1alpha1.StatusRepository
	Addons     []*AddonController
}

// NewAddonsRepository returns pointer to new RepositoryController with url and ready status
func NewAddonsRepository(url string) *RepositoryController {
	return &RepositoryController{
		Repository: addonsv1alpha1.StatusRepository{
			URL:    url,
			Status: addonsv1alpha1.RepositoryStatusReady,
		},
	}
}

// Failed sets StatusRepository as failed
func (ar *RepositoryController) Failed() {
	ar.Repository.Status = addonsv1alpha1.RepositoryStatusFailed
}

// IsFailed checks is StatusRepository is in failed state
func (ar *RepositoryController) IsFailed() bool {
	return ar.Repository.Status == addonsv1alpha1.RepositoryStatusFailed
}
