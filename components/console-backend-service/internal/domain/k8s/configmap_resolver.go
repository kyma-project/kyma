package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=configMapSvc -output=automock -outpkg=automock -case=underscore
type configMapSvc interface {
	Find(name, namespace string) (*v1.ConfigMap, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1.ConfigMap, error)
	Update(name, namespace string, update v1.ConfigMap) (*v1.ConfigMap, error)
	Delete(name, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlConfigMapConverter -output=automock -outpkg=automock -case=underscore
type gqlConfigMapConverter interface {
	ToGQL(in *v1.ConfigMap) (*gqlschema.ConfigMap, error)
	ToGQLs(in []*v1.ConfigMap) ([]gqlschema.ConfigMap, error)
	GQLJSONToConfigMap(in gqlschema.JSON) (v1.ConfigMap, error)
}

type configMapResolver struct {
	configMapSvc       configMapSvc
	configMapConverter gqlConfigMapConverter
}

func newConfigMapResolver(configMapLister configMapSvc) *configMapResolver {
	return &configMapResolver{
		configMapSvc:       configMapLister,
		configMapConverter: &configMapConverter{},
	}
}

func (r *configMapResolver) ConfigMapQuery(ctx context.Context, name, namespace string) (*gqlschema.ConfigMap, error) {
	configMap, err := r.configMapSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	if configMap == nil {
		return nil, nil
	}

	converted, err := r.configMapConverter.ToGQL(configMap)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.ConfigMap, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *configMapResolver) ConfigMapsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ConfigMap, error) {
	configMaps, err := r.configMapSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.ConfigMaps, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMaps, gqlerror.WithNamespace(namespace))
	}

	converted, err := r.configMapConverter.ToGQLs(configMaps)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.ConfigMaps, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMaps, gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *configMapResolver) UpdateConfigMapMutation(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.ConfigMap, error) {
	configMap, err := r.configMapConverter.GQLJSONToConfigMap(update)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace `%s`", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updated, err := r.configMapSvc.Update(name, namespace, configMap)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace %s", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updatedGql, err := r.configMapConverter.ToGQL(updated)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` from namespace %s", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return updatedGql, nil
}

func (r *configMapResolver) DeleteConfigMapMutation(ctx context.Context, name string, namespace string) (*gqlschema.ConfigMap, error) {
	configMap, err := r.configMapSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	configMapCopy := configMap.DeepCopy()
	deletedConfigMap, err := r.configMapConverter.ToGQL(configMapCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ConfigMap))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	err = r.configMapSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.ConfigMap, name, namespace))
		return nil, gqlerror.New(err, pretty.ConfigMap, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedConfigMap, nil
}

func (r *configMapResolver) ConfigMapEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ConfigMapEvent, error) {
	channel := make(chan gqlschema.ConfigMapEvent, 1)
	filter := func(configMap *v1.ConfigMap) bool {
		return configMap != nil && configMap.Namespace == namespace
	}

	configMapListener := listener.NewConfigMap(channel, filter, r.configMapConverter)

	r.configMapSvc.Subscribe(configMapListener)
	go func() {
		defer close(channel)
		defer r.configMapSvc.Unsubscribe(configMapListener)
		<-ctx.Done()
	}()

	return channel, nil
}
