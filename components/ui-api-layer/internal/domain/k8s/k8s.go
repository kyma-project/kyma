package k8s

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
	k8sClientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type PluggableResolver struct {
	params    resolverConstructorParams
	stopCh    chan struct{}
	isEnabled bool
	Resolver
}

func New(restConfig *rest.Config, remoteEnvironmentLister RemoteEnvironmentLister, informerResyncPeriod time.Duration, serviceBindingUsageLister ServiceBindingUsageLister, serviceBindingGetter ServiceBindingGetter) (*PluggableResolver, error) {

	params := resolverConstructorParams{
		restConfig:                restConfig,
		remoteEnvironmentLister:   remoteEnvironmentLister,
		informerResyncPeriod:      informerResyncPeriod,
		serviceBindingUsageLister: serviceBindingUsageLister,
		serviceBindingGetter:      serviceBindingGetter,
	}

	pluggableResolver := &PluggableResolver{
		params:    params,
		isEnabled: false,
		stopCh:    make(chan struct{}),
		Resolver:  &disabled.Resolver{},
	}

	return pluggableResolver, nil
}

// TODO Replace
//go:generate mockery -name=Resolver -output=disabled -outpkg=disabled -case=underscore
type Resolver interface {
	EnvironmentsQuery(ctx context.Context, remoteEnvironment *string) ([]gqlschema.Environment, error)
	SecretQuery(ctx context.Context, name, env string) (*gqlschema.Secret, error)
	DeploymentsQuery(ctx context.Context, environment string, excludeFunctions *bool) ([]gqlschema.Deployment, error)
	DeploymentBoundServiceInstanceNamesField(ctx context.Context, deployment *gqlschema.Deployment) ([]string, error)
	ResourceQuotasQuery(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error)
	ResourceQuotasStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotasStatus, error)
	LimitRangesQuery(ctx context.Context, env string) ([]gqlschema.LimitRange, error)
	InformerFactory() informers.SharedInformerFactory
}

type RemoteEnvironmentLister interface {
	ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error)
	ListNamespacesFor(reName string) ([]string, error)
}

type resolverConstructorParams struct {
	restConfig                *rest.Config
	remoteEnvironmentLister   RemoteEnvironmentLister
	informerResyncPeriod      time.Duration
	serviceBindingUsageLister ServiceBindingUsageLister
	serviceBindingGetter      ServiceBindingGetter
}

func (r *PluggableResolver) Enable() error {
	if r.isEnabled {
		return nil
	}

	r.isEnabled = true
	r.stopCh = make(chan struct{})
	params := r.params

	client, err := v1.NewForConfig(params.restConfig)
	if err != nil {
		return errors.Wrap(err, "while creating K8S Client")
	}

	clientset, err := k8sClientset.NewForConfig(params.restConfig)
	if err != nil {
		return errors.Wrap(err, "while creating K8S Client")
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, params.informerResyncPeriod)

	environmentService := newEnvironmentService(client.Namespaces(), params.remoteEnvironmentLister)
	deploymentService := newDeploymentService(informerFactory.Apps().V1beta2().Deployments().Informer())
	limitRangeService := newLimitRangeService(informerFactory.Core().V1().LimitRanges().Informer())

	resourceQuotaService := newResourceQuotaService(informerFactory.Core().V1().ResourceQuotas().Informer(),
		informerFactory.Apps().V1().ReplicaSets().Informer(), informerFactory.Apps().V1().StatefulSets().Informer(), client)
	resourceQuotaStatusService := newResourceQuotaStatusService(resourceQuotaService, resourceQuotaService, resourceQuotaService, limitRangeService)

	r.Resolver = &k8sResolver{
		environmentResolver:         newEnvironmentResolver(environmentService),
		secretResolver:              newSecretResolver(client),
		deploymentResolver:          newDeploymentResolver(deploymentService, params.serviceBindingUsageLister, params.serviceBindingGetter),
		limitRangeResolver:          newLimitRangeResolver(limitRangeService),
		resourceQuotaResolver:       newResourceQuotaResolver(resourceQuotaService),
		resourceQuotaStatusResolver: newResourceQuotaStatusResolver(resourceQuotaStatusService),
		informerFactory:             informerFactory,
	}

	r.WaitForCacheSync()

	return nil
}

func (r *PluggableResolver) Disable() error {
	if !r.isEnabled {
		return nil
	}

	r.isEnabled = false
	close(r.stopCh)

	// Replace with generated "disabled" k8sResolver
	r.Resolver = &disabled.Resolver{}

	return nil
}

func (r *PluggableResolver) IsEnabled() bool {
	return r.isEnabled
}

func (r *PluggableResolver) Name() string {
	return "k8s"
}

func (r *PluggableResolver) CloseOnKillSignal(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		close(r.stopCh)
	}()
}

func (r *PluggableResolver) WaitForCacheSync() {
	informerFactory := r.InformerFactory()
	informerFactory.Start(r.stopCh)
	informerFactory.WaitForCacheSync(r.stopCh)
}
