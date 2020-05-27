package ui

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterMicroFrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterMicroFrontendConverter interface {
	ToGQL(in *v1alpha1.ClusterMicroFrontend) (*model.ClusterMicroFrontend, error)
	ToGQLs(in []*v1alpha1.ClusterMicroFrontend) ([]*model.ClusterMicroFrontend, error)
}

type clusterMicroFrontendResolver struct {
	service    *resource.Service
	converter *clusterMicroFrontendConverter
}

func newClusterMicroFrontendResolver(sf *resource.ServiceFactory) *clusterMicroFrontendResolver {
	return &clusterMicroFrontendResolver{
		service:    sf.ForResource(v1alpha1.SchemeGroupVersion.WithResource("clustermicrofrontends")),
		converter: newClusterMicroFrontendConverter(),
	}

}

func (r *clusterMicroFrontendResolver) ClusterMicroFrontendsQuery(ctx context.Context) ([]*model.ClusterMicroFrontend, error) {
	//items, err := r.clusterMicroFrontendLister.List()
	list := ClusterMicroFrontendList{}
	err := r.service.List(&list)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterMicroFrontends))
		return nil, gqlerror.New(err, pretty.ClusterMicroFrontends)
	}

	cmfs, err := r.converter.ToGQLs(list)
	if err != nil {
		return nil, err
	}

	return cmfs, nil
}
