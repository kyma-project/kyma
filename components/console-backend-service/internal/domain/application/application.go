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

	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic"
	k8sClient "k8s.io/client-go/kubernetes"

	mappingInformer "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
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
	URL             string
	HTTPCallTimeout time.Duration `envconfig:"default=500ms"`
}

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver               Resolver
	ApplicationRetriever   *applicationRetriever
	mappingInformerFactory mappingInformer.SharedInformerFactory
	appInformerFactory     dynamicinformer.DynamicSharedInformerFactory
	gatewayService         *gateway.Service
}

func New(restConfig *rest.Config, reCfg Config, informerResyncPeriod time.Duration, assetStoreRetriever shared.AssetStoreRetriever, rafterRetriever shared.RafterRetriever) (*PluggableContainer, error) {
	mCli, err := mappingClient.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application broker Clientset")
	}

	aCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application operator Clientset")
	}

	k8sCli, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing application K8s Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			appClient:            aCli,
			mappingClient:        mCli,
			k8sCli:               k8sCli,
			cfg:                  reCfg,
			informerResyncPeriod: informerResyncPeriod,
			assetStoreRetriever:  assetStoreRetriever,
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
	mCli := r.cfg.mappingClient
	aCli := r.cfg.appClient
	kCli := r.cfg.k8sCli

	reCfg := r.cfg.cfg

	// ApplicationMapping
	r.mappingInformerFactory = mappingInformer.NewSharedInformerFactory(mCli, informerResyncPeriod)
	mInformerGroup := r.mappingInformerFactory.Applicationconnector().V1alpha1()
	mInformer := mInformerGroup.ApplicationMappings().Informer()
	mLister := mInformerGroup.ApplicationMappings().Lister()

	// Application
	r.appInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(aCli, informerResyncPeriod)
	aInformer := r.appInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "applications",
	}).Informer()

	aResourceClient := aCli.Resource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "applications",
	})
	appService, err := newApplicationService(reCfg, aResourceClient, mCli.ApplicationconnectorV1alpha1(), mInformer, mLister, aInformer)
	if err != nil {
		return errors.Wrap(err, "while creating Application Service")
	}

	gatewayService, err := gateway.New(kCli, reCfg.Gateway, informerResyncPeriod)
	if err != nil {
		return errors.Wrap(err, "while creating Gateway Service")
	}
	r.gatewayService = gatewayService

	eventActivationService := newEventActivationService(mInformerGroup.EventActivations().Informer())

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.mappingInformerFactory.Start(stopCh)
		r.mappingInformerFactory.WaitForCacheSync(stopCh)

		r.gatewayService.Start(stopCh)

		r.appInformerFactory.Start(stopCh)
		r.appInformerFactory.WaitForCacheSync(stopCh)

		r.Resolver = &domainResolver{
			applicationResolver:     NewApplicationResolver(appService, gatewayService),
			eventActivationResolver: newEventActivationResolver(eventActivationService, r.cfg.assetStoreRetriever, r.cfg.rafterRetriever),
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
	mappingClient        mappingClient.Interface
	appClient            dynamic.Interface
	k8sCli               k8sClient.Interface
	informerResyncPeriod time.Duration
	assetStoreRetriever  shared.AssetStoreRetriever
	rafterRetriever      shared.RafterRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ApplicationQuery(ctx context.Context, name string) (*gqlschema.Application, error)
	ApplicationsQuery(ctx context.Context, namespace *string, first *int, offset *int) ([]gqlschema.Application, error)
	ApplicationEventSubscription(ctx context.Context) (<-chan gqlschema.ApplicationEvent, error)
	CreateApplication(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error)
	DeleteApplication(ctx context.Context, name string) (gqlschema.DeleteApplicationOutput, error)
	UpdateApplication(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error)
	ConnectorServiceQuery(ctx context.Context, application string) (gqlschema.ConnectorService, error)
	EnableApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error)
	OverloadApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error)
	DisableApplicationMutation(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error)
	ApplicationEnabledInNamespacesField(ctx context.Context, obj *gqlschema.Application) ([]string, error)
	ApplicationEnabledMappingServices(ctx context.Context, obj *gqlschema.Application) ([]*gqlschema.EnabledMappingService, error)
	ApplicationStatusField(ctx context.Context, app *gqlschema.Application) (gqlschema.ApplicationStatus, error)
	EventActivationsQuery(ctx context.Context, namespace string) ([]gqlschema.EventActivation, error)
	EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error)
}

type domainResolver struct {
	*applicationResolver
	*eventActivationResolver
}
