package ui

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=clusterMicrofrontendLister -output=automock -outpkg=automock -case=underscore
type clusterMicrofrontendLister interface {
	List() ([]*v1alpha1.ClusterMicroFrontend, error)
}

//go:generate mockery -name=gqlClusterMicrofrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterMicrofrontendConverter interface {
	ToGQL(in *v1alpha1.ClusterMicroFrontend) (*gqlschema.ClusterMicrofrontend, error)
	ToGQLs(in []*v1alpha1.ClusterMicroFrontend) ([]gqlschema.ClusterMicrofrontend, error)
}

type clusterMicrofrontendResolver struct {
	clusterMicrofrontendLister    clusterMicrofrontendLister
	clusterMicrofrontendConverter gqlClusterMicrofrontendConverter
}

func newClusterMicrofrontendResolver(clusterMicrofrontendLister clusterMicrofrontendLister) *clusterMicrofrontendResolver {
	return &clusterMicrofrontendResolver{
		clusterMicrofrontendLister:    clusterMicrofrontendLister,
		clusterMicrofrontendConverter: &clusterMicrofrontendConverter{},
	}
}

func (r *clusterMicrofrontendResolver) ClusterMicrofrontendsQuery(ctx context.Context) ([]gqlschema.ClusterMicrofrontend, error) {
	items, err := r.clusterMicrofrontendLister.List()

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterMicroFrontends))
		return nil, gqlerror.New(err, pretty.ClusterMicroFrontends)
	}

	cmfs, err := r.clusterMicrofrontendConverter.ToGQLs(items)
	if err != nil {
		return nil, err
	}

	return cmfs, nil
}
