package application

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic"
	k8sClient "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/gateway"
)

//go:generate failery -name=ApplicationLister -case=underscore -output disabled -outpkg disabled
type ApplicationLister interface {
	ListInNamespace(namespace string) ([]*v1alpha1.Application, error)
	ListNamespacesFor(appName string) ([]string, error)
}

type applicationRetriever struct {
	ApplicationLister shared.ApplicationLister
}

func (r *applicationRetriever) Application() shared.ApplicationLister {
	return r.ApplicationLister
}

type Config struct {
	Gateway   gateway.Config
	Connector ConnectorSvcCfg
}

type ConnectorSvcCfg struct {
	URL             string        `envconfig:"optional"`
	HTTPCallTimeout time.Duration `envconfig:"default=500ms"`
}

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver               Resolver
	ApplicationRetriever   *applicationRetriever
	mappingInformerFactory dynamicinformer.DynamicSharedInformerFactory
	appInformerFactory     dynamicinformer.DynamicSharedInformerFactory
	gatewayService         *gateway.Service
}

func New(restConfig *rest.Config, reCfg Config, informerResyncPeriod time.Duration, rafterRetriever shared.RafterRetriever) (*PluggableContainer, error) {
	mappingCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application broker Clientset")
	}

	appCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application operator Clientset")
	}

	k8sCli, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application K8s Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			appClient:            appCli,
			mappingClient:        mappingCli,
			k8sCli:               k8sCli,
			cfg:                  reCfg,
			informerResyncPeriod: informerResyncPeriod,
			rafterRetriever:      rafterRetriever,
		},
		Pluggable:            module.NewPluggable("application"),
		ApplicationRetriever: &applicationRetriever{},
	}
	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *PluggableContainer) Enable() error {
	informerResyncPeriod := r.cfg.informerResyncPeriod
	mappingCli := r.cfg.mappingClient
	appCli := r.cfg.appClient
	kCli := r.cfg.k8sCli

	reCfg := r.cfg.cfg

	// Application
	r.appInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(appCli, informerResyncPeriod)
	appInformer := r.appInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "applications",
	}).Informer()

	appClient := appCli.Resource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "applications",
	})

	// ApplicationMapping
	r.mappingInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(mappingCli, informerResyncPeriod)
	mappingInformer := r.mappingInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  mappingTypes.SchemeGroupVersion.Version,
		Group:    mappingTypes.SchemeGroupVersion.Group,
		Resource: "applicationmappings",
	}).Informer()

	mappingClient := mappingCli.Resource(schema.GroupVersionResource{
		Version:  mappingTypes.SchemeGroupVersion.Version,
		Group:    mappingTypes.SchemeGroupVersion.Group,
		Resource: "applicationmappings",
	})

	appService, err := newApplicationService(reCfg, appClient, mappingClient, mappingInformer, appInformer)
	if err != nil {
		return errors.Wrap(err, "while creating Application Service")
	}

	gatewayService, err := gateway.New(kCli, reCfg.Gateway, informerResyncPeriod)
	if err != nil {
		return errors.Wrap(err, "while creating Gateway Service")
	}
	r.gatewayService = gatewayService

	// EventActivations
	eventActivationInformer := r.appInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  mappingTypes.SchemeGroupVersion.Version,
		Group:    mappingTypes.SchemeGroupVersion.Group,
		Resource: "eventactivations",
	}).Informer()

	eventActivationService := newEventActivationService(eventActivationInformer)

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.mappingInformerFactory.Start(stopCh)
		r.mappingInformerFactory.WaitForCacheSync(stopCh)

		r.gatewayService.Start(stopCh)

		r.appInformerFactory.Start(stopCh)
		r.appInformerFactory.WaitForCacheSync(stopCh)

		r.Resolver = &domainResolver{
			applicationResolver:     NewApplicationResolver(appService, gatewayService),
			eventActivationResolver: newEventActivationResolver(eventActivationService, r.cfg.rafterRetriever),
		}
		r.ApplicationRetriever.ApplicationLister = appService
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ApplicationRetriever.ApplicationLister = disabled.NewApplicationLister(disabledErr)
		r.gatewayService = nil
		r.appInformerFactory = nil
		r.mappingInformerFactory = nil
	})

	return nil
}

type resolverConfig struct {
	cfg                  Config
	mappingClient        dynamic.Interface
	appClient            dynamic.Interface
	k8sCli               k8sClient.Interface
	informerResyncPeriod time.Duration
	rafterRetriever      shared.RafterRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ApplicationQuery(ctx context.Context, name string) (*gqlschema.Application, error)
	ApplicationsQuery(ctx context.Context, namespace *string, first *int, offset *int) ([]*gqlschema.Application, error)
	ApplicationEventSubscription(ctx context.Context) (<-chan *gqlschema.ApplicationEvent, error)
	CreateApplication(ctx context.Context, name string, description *string, qglLabels gqlschema.Labels) (*gqlschema.ApplicationMutationOutput, error)
	DeleteApplication(ctx context.Context, name string) (*gqlschema.DeleteApplicationOutput, error)
	UpdateApplication(ctx context.Context, name string, description *string, qglLabels gqlschema.Labels) (*gqlschema.ApplicationMutationOutput, error)
	ConnectorServiceQuery(ctx context.Context, application string) (*gqlschema.ConnectorService, error)
	EnableApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error)
	OverloadApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error)
	DisableApplicationMutation(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error)
	ApplicationEnabledInNamespacesField(ctx context.Context, obj *gqlschema.Application) ([]string, error)
	ApplicationEnabledMappingServices(ctx context.Context, obj *gqlschema.Application) ([]*gqlschema.EnabledMappingService, error)
	ApplicationStatusField(ctx context.Context, app *gqlschema.Application) (gqlschema.ApplicationStatus, error)
	EventActivationsQuery(ctx context.Context, namespace string) ([]*gqlschema.EventActivation, error)
	EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]*gqlschema.EventActivationEvent, error)
}

type domainResolver struct {
	*applicationResolver
	*eventActivationResolver
}
