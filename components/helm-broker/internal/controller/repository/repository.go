package repository

import (
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

type RepositoryController struct {
	Repository addonsv1alpha1.StatusRepository
	Addons     []*AddonController
}

func NewAddonsRepository(url string) *RepositoryController {
	return &RepositoryController{
		Repository: addonsv1alpha1.StatusRepository{
			URL:    url,
			Status: addonsv1alpha1.RepositoryStatusReady,
		},
	}
}

func (ar *RepositoryController) Failed() {
	ar.Repository.Status = addonsv1alpha1.RepositoryStatusFailed
}
