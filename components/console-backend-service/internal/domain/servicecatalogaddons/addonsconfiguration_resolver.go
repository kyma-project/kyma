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

//go:generate mockery -name=addonsCfgService  -output=automock -outpkg=automock -case=underscore
type addonsCfgService interface {
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=addonsCfgLister  -output=automock -outpkg=automock -case=underscore
type addonsCfgLister interface {
	List(namespace string, pagingParams pager.PagingParams) ([]*v1alpha1.AddonsConfiguration, error)
}

//go:generate mockery -name=addonsCfgUpdater  -output=automock -outpkg=automock -case=underscore
type addonsCfgUpdater interface {
	AddRepos(name, namespace string, repository []gqlschema.AddonsConfigurationRepositoryInput) (*v1alpha1.AddonsConfiguration, error)
	RemoveRepos(name, namespace string, reposToRemove []string) (*v1alpha1.AddonsConfiguration, error)
	Resync(name, namespace string) (*v1alpha1.AddonsConfiguration, error)
}

//go:generate mockery -name=addonsCfgMutations  -output=automock -outpkg=automock -case=underscore
type addonsCfgMutations interface {
	Create(name, namespace string, repository []gqlschema.AddonsConfigurationRepositoryInput, labels *gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error)
	Update(name, namespace string, repository []gqlschema.AddonsConfigurationRepositoryInput, labels *gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error)
	Delete(name, namespace string) (*v1alpha1.AddonsConfiguration, error)
}

type addonsConfigurationResolver struct {
	addonsCfgUpdater   addonsCfgUpdater
	addonsCfgLister    addonsCfgLister
	addonsCfgService   addonsCfgService
	addonsCfgMutations addonsCfgMutations
	addonsCfgConverter addonsConfigurationConverter
}

func newAddonsConfigurationResolver(svc *addonsConfigurationService) *addonsConfigurationResolver {
	return &addonsConfigurationResolver{
		addonsCfgLister:    svc,
		addonsCfgUpdater:   svc,
		addonsCfgService:   svc,
		addonsCfgMutations: svc,
		addonsCfgConverter: addonsConfigurationConverter{},
	}
}

func (r *addonsConfigurationResolver) AddonsConfigurationsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error) {
	params := pager.PagingParams{First: first, Offset: offset}

	addons, err := r.addonsCfgLister.List(namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.AddonsConfigurations))
		return nil, gqlerror.New(err, pretty.AddonsConfigurations)
	}

	return r.addonsCfgConverter.ToGQLs(addons), nil
}

func (r *addonsConfigurationResolver) CreateAddonsConfiguration(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(repositories, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}

	addon, err := r.addonsCfgMutations.Create(name, namespace, repositories, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument("urls", strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) UpdateAddonsConfiguration(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(repositories, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}
	addon, err := r.addonsCfgMutations.Update(name, namespace, repositories, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument("urls", strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) DeleteAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgMutations.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) RemoveAddonsConfigurationRepositories(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.RemoveRepos(name, namespace, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while removing repository from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) AddAddonsConfigurationRepositories(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.AddRepos(name, namespace, repositories)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while adding additional repositories to %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

// DEPRECATED: Delete after UI migrate to AddAddonsConfigurationRepositories
func (r *addonsConfigurationResolver) AddAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	repositories, err := resolveRepositories(nil, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resolving repositories from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithDetails(err.Error()))
	}
	addon, err := r.addonsCfgUpdater.AddRepos(name, namespace, repositories)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while adding additional repositories to %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

// DEPRECATED: Delete after UI migrate to RemoveAddonsConfigurationRepositories
func (r *addonsConfigurationResolver) RemoveAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.RemoveRepos(name, namespace, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while removing repository from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) ResyncAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error) {
	addon, err := r.addonsCfgUpdater.Resync(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while resyncing repository from %s %s", pretty.AddonsConfiguration, name))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(addon), nil
}

func (r *addonsConfigurationResolver) AddonsConfigurationEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.AddonsConfigurationEvent, error) {
	channel := make(chan gqlschema.AddonsConfigurationEvent, 1)

	filter := func(entity *v1alpha1.AddonsConfiguration) bool {
		return entity != nil && entity.Namespace == namespace
	}

	brokerListener := listener.NewAddonsConfiguration(channel, filter, &r.addonsCfgConverter)
	r.addonsCfgService.Subscribe(brokerListener)
	go func() {
		defer close(channel)
		defer r.addonsCfgService.Unsubscribe(brokerListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func resolveRepositories(repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string) ([]gqlschema.AddonsConfigurationRepositoryInput, error) {
	if repositories != nil && urls != nil {
		return nil, errors.New("provide only repositories or urls field")
	}
	if repositories == nil {
		repositories = make([]gqlschema.AddonsConfigurationRepositoryInput, 0)
		for _, url := range urls {
			repositories = append(repositories, gqlschema.AddonsConfigurationRepositoryInput{URL: url})
		}
	}
	return repositories, nil
}
