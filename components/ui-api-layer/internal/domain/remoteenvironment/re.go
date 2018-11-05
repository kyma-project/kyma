package remoteenvironment

import (
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type Config struct {
	Gateway   gateway.Config
	Connector ConnectorSvcCfg
}

type ConnectorSvcCfg struct {
	URL             string
	HTTPCallTimeout time.Duration `envconfig:"default=500ms"`
}

type Container struct {
	Resolver *Resolver
	RELister RemoteEnvironmentLister
}

type Resolver struct {
	*remoteEnvironmentResolver
	*eventActivationResolver

	informerFactory externalversions.SharedInformerFactory
	gatewayService  *gateway.Service
}

type RemoteEnvironmentLister interface {
	ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error)
	ListNamespacesFor(reName string) ([]string, error)
}

func New(restConfig *rest.Config, reCfg Config, informerResyncPeriod time.Duration, asyncApiSpecGetter AsyncApiSpecGetter) (*Container, error) {
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Clientset")
	}

	informerFactory := externalversions.NewSharedInformerFactory(client, informerResyncPeriod)
	applicationConnectorGroup := informerFactory.Applicationconnector().V1alpha1()

	envMappingInformer := applicationConnectorGroup.EnvironmentMappings().Informer()
	envMappingLister := applicationConnectorGroup.EnvironmentMappings().Lister()
	remoteEnvInformer := applicationConnectorGroup.RemoteEnvironments().Informer()

	service, err := newRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), reCfg, envMappingInformer, envMappingLister, remoteEnvInformer)
	if err != nil {
		return nil, errors.Wrap(err, "while creating remote environment service")
	}

	gatewayService, err := gateway.New(restConfig, reCfg.Gateway, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while creating gateway service")
	}

	eventActivationService := newEventActivationService(applicationConnectorGroup.EventActivations().Informer())
	return &Container{
		Resolver: &Resolver{
			remoteEnvironmentResolver: NewRemoteEnvironmentResolver(service, gatewayService),
			gatewayService:            gatewayService,
			eventActivationResolver:   newEventActivationResolver(eventActivationService, asyncApiSpecGetter),

			informerFactory: informerFactory,
		},
		RELister: service,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.gatewayService.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
