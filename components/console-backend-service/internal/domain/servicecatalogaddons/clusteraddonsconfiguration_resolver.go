package servicecatalogaddons

import (
	"context"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=clusterAddonsCfgService  -output=automock -outpkg=automock -case=underscore
type clusterAddonsCfgService interface {
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=clusterAddonsCfgLister  -output=automock -outpkg=automock -case=underscore
type clusterAddonsCfgLister interface {
	List(pagingParams pager.PagingParams) ([]*v1alpha1.ClusterAddonsConfiguration, error)
}

//go:generate mockery -name=clusterAddonsCfgUpdater  -output=automock -outpkg=automock -case=underscore
type clusterAddonsCfgUpdater interface {
	AddRepos(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput) (*v1alpha1.ClusterAddonsConfiguration, error)
	RemoveRepos(name string, reposToRemove []string) (*v1alpha1.ClusterAddonsConfiguration, error)
	Resync(name string) (*v1alpha1.ClusterAddonsConfiguration, error)
}

//go:generate mockery -name=clusterAddonsCfgMutations  -output=automock -outpkg=automock -case=underscore
type clusterAddonsCfgMutations interface {
	Create(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error)
	Update(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error)
	Delete(name string) (*v1alpha1.ClusterAddonsConfiguration, error)
}

type clusterAddonsConfigurationResolver struct {
	addonsCfgUpdater   clusterAddonsCfgUpdater
	addonsCfgLister    clusterAddonsCfgLister
	addonsCfgService   clusterAddonsCfgService
	addonsCfgMutations clusterAddonsCfgMutations
	addonsCfgConverter clusterAddonsConfigurationConverter
}

func newClusterAddonsConfigurationResolver(svc *clusterAddonsConfigurationService) *clusterAddonsConfigurationResolver {
	return &clusterAddonsConfigurationResolver{
		addonsCfgLister:    svc,
		addonsCfgUpdater:   svc,
		addonsCfgService:   svc,
		addonsCfgMutations: svc,
		addonsCfgConverter: clusterAddonsConfigurationConverter{},
	}
}

func (r *clusterAddonsConfigurationResolver) ClusterAddonsConfigurationsQuery(ctx context.Context, first *int, offset *int) ([]*gqlschema.AddonsConfiguration, error) {
	params := pager.PagingParams{First: first, Offset: offset}

	addons, err := r.addonsCfgLister.List(params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterAddonsConfigurations))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfigurations)
	}

	return r.addonsCfgConverter.ToGQLs(addons), nil
}

func (r *clusterAddonsConfigurationResolver) CreateClusterAddonsConfiguration(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(repositories, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}

	addon, err := r.addonsCfgMutations.Create(name, repositories, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument("urls", strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) UpdateClusterAddonsConfiguration(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(repositories, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}

	addon, err := r.addonsCfgMutations.Update(name, repositories, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument("urls", strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) DeleteClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgMutations.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) AddClusterAddonsConfigurationRepositories(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.AddRepos(name, repositories)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while adding additional repositories to %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) RemoveClusterAddonsConfigurationRepositories(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.RemoveRepos(name, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while removing repository from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

// DEPRECATED: Delete after UI migrate to AddClusterAddonsConfigurationRepositories
func (r *clusterAddonsConfigurationResolver) AddClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(nil, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}

	addon, err := r.addonsCfgUpdater.AddRepos(name, repositories)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while adding additional repositories to %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

// DEPRECATED: Delete after UI migrate to RemoveClusterAddonsConfigurationRepositories
func (r *clusterAddonsConfigurationResolver) RemoveClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.RemoveRepos(name, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while removing repository from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) ResyncClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.Resync(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resyncing repository from %s %s", pretty.ClusterAddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.ClusterAddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *clusterAddonsConfigurationResolver) ClusterAddonsConfigurationEventSubscription(ctx context.Context) (<-chan *gqlschema.ClusterAddonsConfigurationEvent, error) {
	channel := make(chan *gqlschema.ClusterAddonsConfigurationEvent, 1)

	filter := func(entity *v1alpha1.ClusterAddonsConfiguration) bool {
		return entity != nil
	}

	brokerListener := listener.NewClusterAddonsConfiguration(channel, filter, &r.addonsCfgConverter)
	r.addonsCfgService.Subscribe(brokerListener)
	go func() {
		defer close(channel)
		defer r.addonsCfgService.Unsubscribe(brokerListener)
		<-ctx.Done()
	}()

	return channel, nil
}
