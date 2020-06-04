package kymahelm

import (
	"log"
	"strings"
	"time"
)

const (
	defaultMaxIterations = 10
	defaultSleepTimeSec  = 10
)

type WaitReleaseStatusFunc func(releaseName, releaseNamespace string) (ReleaseStatus, error)
type WaitPredicateFunc func(releaseStatus ReleaseStatus, relStatusResponseErr error) (bool, error)

type WaitOption func(*waitConditionCfg)

type waitConditionCfg struct {
	releaseStatusFn WaitReleaseStatusFunc
	predicateFn     WaitPredicateFunc
	sleepTimeSec    uint8
	maxIterations   uint8
}

func (wc *waitConditionCfg) wait(releaseName, releaseNamespace string) (bool, error) {

	var fulfilled bool
	var iter uint8 = 0

	for !fulfilled && iter < wc.maxIterations {

		relStatus, relStatusErr := wc.releaseStatusFn(releaseName, releaseNamespace)

		eval, predicateError := wc.predicateFn(relStatus, relStatusErr)

		if predicateError != nil {
			return false, predicateError
		}

		fulfilled = eval
		if fulfilled {
			return true, nil
		}

		time.Sleep(time.Duration(wc.sleepTimeSec) * time.Second)
		iter += 1
	}

	return fulfilled, nil
}

// WaitForCondition returns true if condition was fulfilled within configured time, returns false otherwise.
// Returns an error immediately if the configured predicate function returns an error.
func (hc *Client) WaitForCondition(releaseName string, releaseNamespace string, pf WaitPredicateFunc, opts ...WaitOption) (bool, error) {

	//default
	relStatusFn := func(releaseName, releaseNamespace string) (ReleaseStatus, error) {
		//No retries here on purpose. Perhaps the user wants to wait for "release not exist" condition. That will return an error.
		//Implement smarter error handling/retries with the help of WaitPredicateFunc
		_, err := hc.ReleaseStatus(releaseName, releaseNamespace) //TODO: Improve
		return ReleaseStatus{}, err
	}

	cfg := defaultWaitConditionCfg()
	cfg.releaseStatusFn = relStatusFn
	cfg.predicateFn = pf

	//user-provided opts
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg.wait(releaseName, releaseNamespace)
}

func (hc *Client) WaitForReleaseDelete(releaseName, releaseNamespace string) (bool, error) {

	pf := func(relStatus ReleaseStatus, getStatusRespErr error) (bool, error) {
		if getStatusRespErr != nil {
			if strings.Contains(getStatusRespErr.Error(), "release Not Found") { //TODO: Improve!
				return true, nil
			}
			log.Printf("Error while waiting for release delete: %s", getStatusRespErr.Error())
			//Continue waiting
			return false, nil
		}

		log.Printf("Waiting for release delete: release status: %s/%s: %s", releaseNamespace, releaseName, relStatus.Status)

		//Continue waiting
		return false, nil
	}

	return hc.WaitForCondition(releaseName, releaseNamespace, pf)
}

func (hc *Client) WaitForReleaseRollback(releaseName, releaseNamespace string) (bool, error) {

	pf := func(releaseStatus ReleaseStatus, getStatusRespErr error) (bool, error) {
		if getStatusRespErr != nil {
			log.Printf("Error while waiting for release rollback: %s", getStatusRespErr.Error())
			//Continue waiting
			return false, nil
		}

		if releaseStatus.Status == "DEPLOYED" { //TODO: Improve
			return true, nil
		}
		log.Printf("Waiting for release rollback: release status: %s/%s: %s", releaseNamespace, releaseName, releaseStatus.Status)

		//Continue waiting
		return false, nil
	}

	return hc.WaitForCondition(releaseName, releaseNamespace, pf)
}
func defaultWaitConditionCfg() *waitConditionCfg {

	return &waitConditionCfg{
		sleepTimeSec:  defaultSleepTimeSec,
		maxIterations: defaultMaxIterations,
	}
}

func ReleaseStatusFunc(val WaitReleaseStatusFunc) WaitOption {
	return func(wcfg *waitConditionCfg) {
		wcfg.releaseStatusFn = val
	}
}

func SleepTimeSecs(val uint8) WaitOption {
	return func(wcfg *waitConditionCfg) {
		wcfg.sleepTimeSec = val
	}
}

func MaxIterations(val uint8) WaitOption {
	return func(wcfg *waitConditionCfg) {
		wcfg.maxIterations = val
	}
}
