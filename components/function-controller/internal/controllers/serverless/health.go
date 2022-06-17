package serverless

import (
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

// This const should be longer than 253 characters to avoid collisions with real k8s objects.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
// This event is artificial and it's only used to check if reconciliation loop didn't stop reconciling
// The event is not fully validated, that's why we can use invalid name.
const HealthEvent = "HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT"

var _ healthz.Checker = HealthChecker{}.Checker

type HealthChecker struct {
	checkCh  chan event.GenericEvent
	healthCh chan bool
	timeout  time.Duration
	log      *zap.SugaredLogger
}

func NewHealthChecker(checkCh chan event.GenericEvent, returnCh chan bool, timeout time.Duration, logger *zap.SugaredLogger) HealthChecker {
	return HealthChecker{checkCh: checkCh, healthCh: returnCh, timeout: timeout, log: logger}
}

func (h HealthChecker) Checker(req *http.Request) error {
	h.log.Debug("Liveness handler triggered")

	checkEvent := event.GenericEvent{
		Object: &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name: HealthEvent,
			},
		},
	}
	select {
	case h.checkCh <- checkEvent:
	case <-time.After(h.timeout):
		return errors.New("timeout when sending check event")
	}

	h.log.Debug("check event send to reconcile loop")
	select {
	case <-h.healthCh:
		h.log.Debug("reconcile loop is healthy")
		return nil
	case <-time.After(h.timeout):
		h.log.Debug("reconcile timeout")
		return errors.New("reconcile didn't send confirmation")
	}
}

func IsHealthCheckRequest(req ctrl.Request) bool {
	return req.Name == HealthEvent
}
