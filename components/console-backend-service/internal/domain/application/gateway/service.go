package gateway

import (
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/executor"
	"k8s.io/client-go/kubernetes"
)

// Service gives a functionality to provide gateway status. It hides implementation details
// and exports necessary methods.
type Service struct {
	provider      *provider
	statusWatcher *gatewayStatusWatcher

	cfg Config
}

func New(k8sCli kubernetes.Interface, reCfg Config, informerResyncPeriod time.Duration) (*Service, error) {
	gatewaySvcProvider := newProvider(k8sCli.CoreV1(), reCfg.IntegrationNamespace, informerResyncPeriod)
	watcher := newStatusWatcher(gatewaySvcProvider, reCfg.StatusCallTimeout)

	return &Service{
		provider:      gatewaySvcProvider,
		statusWatcher: watcher,
		cfg:           reCfg,
	}, nil
}

func (svc *Service) Start(stopCh <-chan struct{}) {
	svc.provider.WaitForCacheSync(stopCh)
	executor.NewPeriodic(svc.cfg.StatusRefreshPeriod, func(stopCh <-chan struct{}) {
		svc.statusWatcher.Refresh(stopCh)
	}).Run(stopCh)
}

func (svc *Service) GetStatus(appName string) Status {
	return svc.statusWatcher.GetStatus(appName)
}
