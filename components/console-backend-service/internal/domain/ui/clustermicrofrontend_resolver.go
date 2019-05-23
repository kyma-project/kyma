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

//go:generate mockery -name=clusterMicroFrontendLister -output=automock -outpkg=automock -case=underscore
type clusterMicroFrontendLister interface {
	List() ([]*v1alpha1.ClusterMicroFrontend, error)
}

//go:generate mockery -name=gqlClusterMicroFrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterMicroFrontendConverter interface {
	ToGQL(in *v1alpha1.ClusterMicroFrontend) (*gqlschema.ClusterMicroFrontend, error)
	ToGQLs(in []*v1alpha1.ClusterMicroFrontend) ([]gqlschema.ClusterMicroFrontend, error)
}

type clusterMicroFrontendResolver struct {
	clusterMicroFrontendLister    clusterMicroFrontendLister
	clusterMicroFrontendConverter gqlClusterMicroFrontendConverter
}

func newClusterMicroFrontendResolver(clusterMicroFrontendLister clusterMicroFrontendLister) *clusterMicroFrontendResolver {
	return &clusterMicroFrontendResolver{
		clusterMicroFrontendLister:    clusterMicroFrontendLister,
		clusterMicroFrontendConverter: newClusterMicroFrontendConverter(),
	}
}

func (r *clusterMicroFrontendResolver) ClusterMicroFrontendsQuery(ctx context.Context) ([]gqlschema.ClusterMicroFrontend, error) {
	items, err := r.clusterMicroFrontendLister.List()

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterMicroFrontends))
		return nil, gqlerror.New(err, pretty.ClusterMicroFrontends)
	}

	cmfs, err := r.clusterMicroFrontendConverter.ToGQLs(items)
	if err != nil {
		return nil, err
	}

	return cmfs, nil
}
