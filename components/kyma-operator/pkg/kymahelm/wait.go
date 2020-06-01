package kymahelm

import (
	"log"
	"strings"
	"time"

	rls "k8s.io/helm/pkg/proto/hapi/services"
	helmerrors "k8s.io/helm/pkg/storage/errors"
)

type WaitReleaseStatusFunc func(releaseName string) (*rls.GetReleaseStatusResponse, error)
type WaitPredicateFunc func(relStatusResponse *rls.GetReleaseStatusResponse, relStatusResponseErr error) (bool, error)

type WaitOption func(*waitConditionCfg)

type waitConditionCfg struct {
	releaseStatusFn WaitReleaseStatusFunc
	predicateFn     WaitPredicateFunc
	sleepTimeSec    uint8
	maxIterations   uint8
}

func (wc *waitConditionCfg) wait(releaseName string) (bool, error) {

	var fulfilled bool
	var iter uint8 = 0

	for !fulfilled && iter < wc.maxIterations {

		status, relStatusErr := wc.releaseStatusFn(releaseName)

		eval, predicateError := wc.predicateFn(status, relStatusErr)

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
func (hc *Client) WaitForCondition(releaseName string, pf WaitPredicateFunc, opts ...WaitOption) (bool, error) {

	//default
	relStatusFn := func(releaseName string) (*rls.GetReleaseStatusResponse, error) {
		//No retries here on purpose. Perhaps the user wants to wait for "release not exist" condition. That will return an error.
		//Implement smarter error handling/retries with the help of WaitPredicateFunc
		return hc.helm.ReleaseStatus(releaseName)
	}

	cfg := defaultWaitConditionCfg()
	cfg.releaseStatusFn = relStatusFn
	cfg.predicateFn = pf

	//user-provided opts
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg.wait(releaseName)
}

func (hc *Client) WaitForReleaseDelete(releaseName string) (bool, error) {

	pf := func(resp *rls.GetReleaseStatusResponse, getStatusRespErr error) (bool, error) {
		if getStatusRespErr != nil {
			if strings.Contains(getStatusRespErr.Error(), helmerrors.ErrReleaseNotFound(releaseName).Error()) {
				return true, nil
			}
			log.Printf("Error while waiting for release delete: %s", getStatusRespErr.Error())
			//Continue waiting
			return false, nil
		}

		if resp != nil {
			log.Printf("Waiting for release delete: release status: %s/%s: %s", resp.Namespace, resp.Name, resp.Info.Status.Code.String())
		}

		//Continue waiting
		return false, nil
	}

	return hc.WaitForCondition(releaseName, pf)
}

func (hc *Client) WaitForReleaseRollback(releaseName string) (bool, error) {

	pf := func(resp *rls.GetReleaseStatusResponse, getStatusRespErr error) (bool, error) {
		if getStatusRespErr != nil {
			log.Printf("Error while waiting for release rollback: %s", getStatusRespErr.Error())
			//Continue waiting
			return false, nil
		}

		if resp != nil {
			if resp.Info.Status.Code.String() == "DEPLOYED" {
				return true, nil
			}
			log.Printf("Waiting for release rollback: release status: %s/%s: %s", resp.Namespace, resp.Name, resp.Info.Status.Code.String())
		}

		//Continue waiting
		return false, nil
	}

	return hc.WaitForCondition(releaseName, pf)
}
func defaultWaitConditionCfg() *waitConditionCfg {

	return &waitConditionCfg{
		sleepTimeSec:  10,
		maxIterations: 10,
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
