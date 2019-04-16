package servicecatalogaddons

import (
	"context"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=addonsCfgService  -output=automock -outpkg=automock -case=underscore
type addonsCfgService interface {
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=addonsCfgLister  -output=automock -outpkg=automock -case=underscore
type addonsCfgLister interface {
	List(pagingParams pager.PagingParams) ([]*v1.ConfigMap, error)
}

//go:generate mockery -name=addonsCfgUpdater  -output=automock -outpkg=automock -case=underscore
type addonsCfgUpdater interface {
	AddRepos(name string, url []string) (*v1.ConfigMap, error)
	RemoveRepos(name string, urls []string) (*v1.ConfigMap, error)
}

//go:generate mockery -name=addonsCfgMutations  -output=automock -outpkg=automock -case=underscore
type addonsCfgMutations interface {
	Create(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error)
	Update(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error)
	Delete(name string) (*v1.ConfigMap, error)
}

type addonsConfigurationResolver struct {
	addonsCfgUpdater   addonsCfgUpdater
	addonsCfgLister    addonsCfgLister
	addonsCfgService   addonsCfgService
	addonsCfgMutations addonsCfgMutations
	addonsCfgConverter addonsConfigurationConverter
}

func newAddonsRepoResolver(svc *addonsConfigurationService) *addonsConfigurationResolver {
	return &addonsConfigurationResolver{
		addonsCfgLister:    svc,
		addonsCfgUpdater:   svc,
		addonsCfgService:   svc,
		addonsCfgMutations: svc,
		addonsCfgConverter: addonsConfigurationConverter{},
	}
}

func (r *addonsConfigurationResolver) AddonsConfigurationsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error) {
	cfgs, err := r.addonsCfgLister.List(pager.PagingParams{First: first, Offset: offset})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.AddonsConfigurations))
		return nil, gqlerror.New(err, pretty.AddonsConfigurations)
	}

	return r.addonsCfgConverter.ToGQLs(cfgs), nil
}

func (r *addonsConfigurationResolver) CreateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	cfg, err := r.addonsCfgMutations.Create(name, urls, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s", pretty.AddonsConfiguration))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument(addonsCfgKey, strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(cfg), nil
}

func (r *addonsConfigurationResolver) UpdateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	cfg, err := r.addonsCfgMutations.Update(name, urls, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s", pretty.AddonsConfiguration))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name), gqlerror.WithCustomArgument(addonsCfgKey, strings.Join(urls, "\n")))
	}

	return r.addonsCfgConverter.ToGQL(cfg), nil
}

func (r *addonsConfigurationResolver) DeleteAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	cfg, err := r.addonsCfgMutations.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s", pretty.AddonsConfiguration))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(cfg), nil
}

func (r addonsConfigurationResolver) AddAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	cfg, err := r.addonsCfgUpdater.AddRepos(name, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while adding additional repositories to %s", pretty.AddonsConfiguration))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(cfg), nil
}

func (r *addonsConfigurationResolver) RemoveAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	cfg, err := r.addonsCfgUpdater.RemoveRepos(name, urls)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while removing repository from %s", pretty.AddonsConfiguration))
		return nil, gqlerror.New(err, pretty.AddonsConfiguration, gqlerror.WithName(name))
	}

	return r.addonsCfgConverter.ToGQL(cfg), nil
}

func (r *addonsConfigurationResolver) AddonsConfigurationEventSubscription(ctx context.Context) (<-chan gqlschema.AddonsConfigurationEvent, error) {
	channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
	filter := func(entity *v1.ConfigMap) bool {
		if entity == nil {
			return false
		}
		if entity.Namespace != systemNs {
			return false
		}
		val, ok := entity.Labels[addonsCfgLabelKey]
		if !ok {
			return false
		}
		if val != addonsCfgLabelValue {
			return false
		}
		return true
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
