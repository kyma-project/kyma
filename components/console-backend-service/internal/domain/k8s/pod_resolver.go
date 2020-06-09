package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=podSvc -output=automock -outpkg=automock -case=underscore
type podSvc interface {
	Find(name, namespace string) (*v1.Pod, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1.Pod, error)
	Update(name, namespace string, update v1.Pod) (*v1.Pod, error)
	Delete(name, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlPodConverter -output=automock -outpkg=automock -case=underscore
type gqlPodConverter interface {
	ToGQL(in *v1.Pod) (*gqlschema.Pod, error)
	ToGQLs(in []*v1.Pod) ([]*gqlschema.Pod, error)
	GQLJSONToPod(in gqlschema.JSON) (v1.Pod, error)
}

type podResolver struct {
	podSvc       podSvc
	podConverter gqlPodConverter
}

func newPodResolver(podSvc podSvc) *podResolver {
	return &podResolver{
		podSvc:       podSvc,
		podConverter: &podConverter{},
	}
}

func (r *podResolver) PodQuery(ctx context.Context, name, namespace string) (*gqlschema.Pod, error) {
	pod, err := r.podSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	if pod == nil {
		return nil, nil
	}

	converted, err := r.podConverter.ToGQL(pod)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.Pod, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *podResolver) PodsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.Pod, error) {
	pods, err := r.podSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.Pods, namespace))
		return nil, gqlerror.New(err, pretty.Pods, gqlerror.WithNamespace(namespace))
	}

	converted, err := r.podConverter.ToGQLs(pods)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.Pods, namespace))
		return nil, gqlerror.New(err, pretty.Pods, gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *podResolver) PodEventSubscription(ctx context.Context, namespace string) (<-chan *gqlschema.PodEvent, error) {
	channel := make(chan *gqlschema.PodEvent, 1)
	filter := func(pod *v1.Pod) bool {
		return pod != nil && pod.Namespace == namespace
	}

	podListener := listener.NewPod(channel, filter, r.podConverter)

	r.podSvc.Subscribe(podListener)
	go func() {
		defer close(channel)
		defer r.podSvc.Unsubscribe(podListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *podResolver) UpdatePodMutation(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.Pod, error) {
	pod, err := r.podConverter.GQLJSONToPod(update)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace `%s`", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updated, err := r.podSvc.Update(name, namespace, pod)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace %s", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updatedGql, err := r.podConverter.ToGQL(updated)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` from namespace %s", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return updatedGql, nil
}

func (r *podResolver) DeletePodMutation(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	pod, err := r.podSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	podCopy := pod.DeepCopy()
	deletedPod, err := r.podConverter.ToGQL(podCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Pod))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	err = r.podSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.Pod, name, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedPod, nil
}
