package application

import (
	"time"

	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	mappingInformer "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	appInformer "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/gateway"
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
	Resolver  *Resolver
	AppLister ApplicationLister
}

type Resolver struct {
	*applicationResolver
	*eventActivationResolver

	mappingInformerFactory mappingInformer.SharedInformerFactory
	appInformerFactory     appInformer.SharedInformerFactory
	gatewayService         *gateway.Service
}

type ApplicationLister interface {
	ListInEnvironment(environment string) ([]*v1alpha1.Application, error)
	ListNamespacesFor(appName string) ([]string, error)
}

func New(restConfig *rest.Config, reCfg Config, informerResyncPeriod time.Duration, asyncApiSpecGetter AsyncApiSpecGetter) (*Container, error) {
	// ApplicationMapping
	mCli, err := mappingClient.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application broker Clientset")
	}
	mInformerFactory := mappingInformer.NewSharedInformerFactory(mCli, informerResyncPeriod)
	mInformerGroup := mInformerFactory.Applicationconnector().V1alpha1()

	mInformer := mInformerGroup.ApplicationMappings().Informer()
	mLister := mInformerGroup.ApplicationMappings().Lister()

	// Application
	aCli, err := appCli.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application operator Clientset")
	}

	aInformerFactory := appInformer.NewSharedInformerFactory(aCli, informerResyncPeriod)
	aInformer := aInformerFactory.Applicationconnector().V1alpha1().Applications().Informer()

	// Service
	service, err := newApplicationService(reCfg, aCli.ApplicationconnectorV1alpha1(), mCli.ApplicationconnectorV1alpha1(), mInformer, mLister, aInformer)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Application service")
	}

	gatewayService, err := gateway.New(restConfig, reCfg.Gateway, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while creating gateway service")
	}

	eventActivationService := newEventActivationService(mInformerGroup.EventActivations().Informer())
	return &Container{
		Resolver: &Resolver{
			applicationResolver:     NewApplicationResolver(service, gatewayService),
			gatewayService:          gatewayService,
			eventActivationResolver: newEventActivationResolver(eventActivationService, asyncApiSpecGetter),

			mappingInformerFactory: mInformerFactory,
			appInformerFactory:     aInformerFactory,
		},
		AppLister: service,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.mappingInformerFactory.Start(stopCh)
	r.appInformerFactory.Start(stopCh)

	r.gatewayService.Start(stopCh)

	r.mappingInformerFactory.WaitForCacheSync(stopCh)
	r.appInformerFactory.WaitForCacheSync(stopCh)
}
