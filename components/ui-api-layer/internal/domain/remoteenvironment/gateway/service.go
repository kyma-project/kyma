package gateway

import (
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/executor"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Service gives a functionality to provide gateway status. It hides implementation details
// and exports necessary methods.
type Service struct {
	provider      *provider
	statusWatcher *gatewayStatusWatcher

	cfg Config
}

func New(restConfig *rest.Config, reCfg Config, informerResyncPeriod time.Duration) (*Service, error) {
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Clientset")
	}
	gatewaySvcProvider := newProvider(k8sClient.CoreV1(), reCfg.IntegrationNamespace, informerResyncPeriod)
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

func (svc *Service) GetStatus(reName string) Status {
	return svc.statusWatcher.GetStatus(reName)
}
