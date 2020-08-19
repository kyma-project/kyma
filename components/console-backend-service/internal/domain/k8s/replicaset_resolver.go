package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	api "k8s.io/api/apps/v1"
)

//go:generate mockery -name=replicaSetSvc -output=automock -outpkg=automock -case=underscore
type replicaSetSvc interface {
	Find(name, namespace string) (*api.ReplicaSet, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*api.ReplicaSet, error)
	Update(name, namespace string, update api.ReplicaSet) (*api.ReplicaSet, error)
	Delete(name, namespace string) error
}

//go:generate mockery -name=gqlReplicaSetConverter -output=automock -outpkg=automock -case=underscore
type gqlReplicaSetConverter interface {
	ToGQL(in *api.ReplicaSet) (*gqlschema.ReplicaSet, error)
	ToGQLs(in []*api.ReplicaSet) ([]*gqlschema.ReplicaSet, error)
	GQLJSONToReplicaSet(in gqlschema.JSON) (api.ReplicaSet, error)
}

type replicaSetResolver struct {
	replicaSetSvc       replicaSetSvc
	replicaSetConverter gqlReplicaSetConverter
}

func newReplicaSetResolver(replicaSetLister replicaSetSvc) *replicaSetResolver {
	return &replicaSetResolver{
		replicaSetSvc:       replicaSetLister,
		replicaSetConverter: &replicaSetConverter{},
	}
}

func (r *replicaSetResolver) ReplicaSetQuery(ctx context.Context, name, namespace string) (*gqlschema.ReplicaSet, error) {
	replicaSet, err := r.replicaSetSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	if replicaSet == nil {
		return nil, nil
	}

	converted, err := r.replicaSetConverter.ToGQL(replicaSet)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.ReplicaSet, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *replicaSetResolver) ReplicaSetsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.ReplicaSet, error) {
	replicaSets, err := r.replicaSetSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.ReplicaSets, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSets, gqlerror.WithNamespace(namespace))
	}

	converted, err := r.replicaSetConverter.ToGQLs(replicaSets)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s from namespace %s", pretty.ReplicaSets, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSets, gqlerror.WithNamespace(namespace))
	}

	return converted, nil
}

func (r *replicaSetResolver) UpdateReplicaSetMutation(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.ReplicaSet, error) {
	replicaSet, err := r.replicaSetConverter.GQLJSONToReplicaSet(update)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace `%s`", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updated, err := r.replicaSetSvc.Update(name, namespace, replicaSet)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace %s", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updatedGql, err := r.replicaSetConverter.ToGQL(updated)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` from namespace %s", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return updatedGql, nil
}

func (r *replicaSetResolver) DeleteReplicaSetMutation(ctx context.Context, name string, namespace string) (*gqlschema.ReplicaSet, error) {
	replicaSet, err := r.replicaSetSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	replicaSetCopy := replicaSet.DeepCopy()
	deletedReplicaSet, err := r.replicaSetConverter.ToGQL(replicaSetCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ReplicaSet))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	err = r.replicaSetSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.ReplicaSet, name, namespace))
		return nil, gqlerror.New(err, pretty.ReplicaSet, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedReplicaSet, nil
}
