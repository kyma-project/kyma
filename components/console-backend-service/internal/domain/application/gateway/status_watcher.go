package gateway

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/iosafety"
)

type Status string

const (
	StatusNotServing    Status = "NotServing"
	StatusServing       Status = "Serving"
	StatusNotConfigured Status = "GatewayNotConfigured"
)

//go:generate mockery -name=gatewayServiceLister -output=automock -outpkg=automock -case=underscore
type gatewayServiceLister interface {
	ListGatewayServices() []ServiceData
}

type gatewayStatusWatcher struct {
	gatewayServiceLister gatewayServiceLister
	healthiness          map[string]bool

	mu          sync.RWMutex
	httpTimeout time.Duration

	httpClient *http.Client
}

func newStatusWatcher(gatewayServiceLister gatewayServiceLister, httpTimeout time.Duration) *gatewayStatusWatcher {
	httpClient := &http.Client{
		Timeout: httpTimeout,
	}
	return &gatewayStatusWatcher{
		gatewayServiceLister: gatewayServiceLister,
		httpTimeout:          httpTimeout,
		healthiness:          map[string]bool{},
		httpClient:           httpClient,
	}
}

// GetStatus returns status of the Application
func (s *gatewayStatusWatcher) GetStatus(appName string) Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	healthy, exists := s.healthiness[appName]
	if !exists {
		return StatusNotConfigured
	}

	if !healthy {
		return StatusNotServing
	}

	return StatusServing
}

func (s *gatewayStatusWatcher) Refresh(stopCh <-chan struct{}) {
	items := s.gatewayServiceLister.ListGatewayServices()

	localHealthiness := map[string]bool{}
	for _, item := range items {
		select {
		case <-stopCh:
			return
		default:
		}

		healthy, err := s.isHealthy(fmt.Sprintf("http://%s/v1/health", item.Host))
		if err != nil {
			//glog.Warningf("Application %s health check failed (%s), error: %s",
			//	item.ApplicationName, item.Host, err.Error())
		}
		localHealthiness[item.ApplicationName] = healthy
	}

	s.swapStatusData(localHealthiness)
}

func (s *gatewayStatusWatcher) swapStatusData(localStatusMap map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.healthiness = localStatusMap
}

func (s *gatewayStatusWatcher) isHealthy(url string) (bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		err := iosafety.DrainReader(resp.Body)
		if err != nil {
			glog.Errorf("Unable to drain body reader. Cause: %v", err)
		}
		err = resp.Body.Close()
		if err != nil {
			glog.Errorf("Unable to close body reader. Cause: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("expect HTTP status code %d got HTTP status code %d", http.StatusOK, resp.StatusCode)
	}

	return true, nil
}
