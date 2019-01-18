package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=podLister -output=automock -outpkg=automock -case=underscore
type podLister interface {
	Find(name, namespace string) (*v1.Pod, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1.Pod, error)
}

//go:generate mockery -name=gqlPodConverter -output=automock -outpkg=automock -case=underscore
type gqlPodConverter interface {
	ToGQL(in *v1.Pod) (*gqlschema.Pod, error)
	ToGQLs(in []*v1.Pod) ([]gqlschema.Pod, error)
}

type podResolver struct {
	podLister    podLister
	podConverter gqlPodConverter
}

func newPodResolver(podLister podLister) *podResolver {
	return &podResolver{
		podLister:    podLister,
		podConverter: &podConverter{},
	}
}

func (r *podResolver) PodQuery(ctx context.Context, name, namespace string) (*gqlschema.Pod, error) {
	pod, err := r.podLister.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithEnvironment(namespace))
	}
	if pod == nil {
		return nil, nil
	}

	converted, err := r.podConverter.ToGQL(pod)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.Pod, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithEnvironment(namespace))
	}

	return converted, nil
}

func (r *podResolver) PodsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Pod, error) {
	pods, err := r.podLister.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.Pods, namespace))
		return nil, gqlerror.New(err, pretty.Pods, gqlerror.WithEnvironment(namespace))
	}

	converted, err := r.podConverter.ToGQLs(pods)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.Pods, namespace))
		return nil, gqlerror.New(err, pretty.Pods, gqlerror.WithEnvironment(namespace))
	}

	return converted, nil
}
