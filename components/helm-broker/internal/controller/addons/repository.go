package addons

import (
	"fmt"

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

// HasFailedAddons returns true if any addon in the repository has status Failed
func (ar *RepositoryController) HasFailedAddons() bool {
	for _, addon := range ar.Addons {
		if !addon.IsReady() {
			return true
		}
	}
	return false
}

// FetchingError sets StatusRepository as failed with URLFetchingError as a reason
func (ar *RepositoryController) FetchingError(err error) {
	reason := addonsv1alpha1.RepositoryURLFetchingError
	ar.Failed()
	ar.Repository.Reason = reason
	ar.Repository.Message = fmt.Sprintf(reason.Message(), err.Error())
}
