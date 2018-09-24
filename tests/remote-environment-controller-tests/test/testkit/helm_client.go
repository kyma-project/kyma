package testkit

import (
	"k8s.io/helm/pkg/helm"
	"time"
)

type HelmClient interface {
	ExistWhenShould(releaseName string) (bool, error)
	ExistWhenShouldNot(releaseName string) (bool, error)
}

type helmClient struct {
	helm          *helm.Client
	retryCount    int
	retryWaitTime time.Duration
}

func NewHelmClient(host string, retryCount int, retryWaitTime time.Duration) HelmClient {
	return &helmClient{
		helm:          helm.NewClient(helm.Host(host)),
		retryCount:    retryCount,
		retryWaitTime: retryWaitTime,
	}
}

func (hc *helmClient) ExistWhenShould(releaseName string) (bool, error) {
	return hc.checkExistenceWithRetriesIf(releaseName, shouldRetryIfNotExist)
}

func (hc *helmClient) ExistWhenShouldNot(releaseName string) (bool, error) {
	return hc.checkExistenceWithRetriesIf(releaseName, shouldRetryIfExists)
}

func (hc *helmClient) checkExistenceWithRetriesIf(releaseName string, shouldRetry func(releaseExists bool, err error) bool) (bool, error) {
	var exists bool
	var err error

	for i := 0; i < hc.retryCount && shouldRetry(exists, err); i++ {
		exists, err = hc.checkReleaseExistence(releaseName)
		if shouldRetry(exists, err) {
			time.Sleep(hc.retryWaitTime)
		}
	}

	return exists, err
}

func shouldRetryIfNotExist(releaseExists bool, err error) bool {
	return err != nil || releaseExists == false
}

func shouldRetryIfExists(releaseExists bool, err error) bool {
	return err != nil || releaseExists == true
}

func (hc *helmClient) checkReleaseExistence(name string) (bool, error) {
	listResponse, err := hc.helm.ListReleases()
	if err != nil {
		return false, err
	}
	releases := listResponse.Releases
	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}
	return false, nil
}
