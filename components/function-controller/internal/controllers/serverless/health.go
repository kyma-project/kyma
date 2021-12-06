package serverless

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

// This const is special now valid k8s resource name to avoid conflict with real resource name
const HEALTH_EVENT = "HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT"

var _ healthz.Checker = HealthChecker{}.Checker

type HealthChecker struct {
	checkCh  chan event.GenericEvent
	healthCh chan bool
	timeout  time.Duration
}

func NewChecker(checkCh chan event.GenericEvent, returnCh chan bool, timeout time.Duration) HealthChecker {
	return HealthChecker{checkCh: checkCh, healthCh: returnCh, timeout: timeout}
}

func (h HealthChecker) Checker(req *http.Request) error {

	checkEvent := event.GenericEvent{Meta: &ctrl.ObjectMeta{Name: HEALTH_EVENT}}
	select {
	case h.checkCh <- checkEvent:

	case <-time.After(h.timeout):
		return errors.New("timeout when sending check event")
	}

	fmt.Println("before receiving")
	select {
	case <-h.healthCh:
		fmt.Println("received")

		return nil
	case <-time.After(h.timeout):
		return errors.New("reconcile didn't send confirmation")
	}
}

func IsHealthCheckRequest(req ctrl.Request) bool {
	return req.Name == HEALTH_EVENT
}
