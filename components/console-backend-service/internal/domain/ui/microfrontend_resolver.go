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

//go:generate mockery -name=microFrontendLister -output=automock -outpkg=automock -case=underscore
type microFrontendLister interface {
	List(namespace string) ([]*v1alpha1.MicroFrontend, error)
}

//go:generate mockery -name=gqlMicroFrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlMicroFrontendConverter interface {
	ToGQL(in *v1alpha1.MicroFrontend) (*gqlschema.MicroFrontend, error)
	ToGQLs(in []*v1alpha1.MicroFrontend) ([]gqlschema.MicroFrontend, error)
}

type microFrontendResolver struct {
	microFrontendLister    microFrontendLister
	microFrontendConverter gqlMicroFrontendConverter
}

func newMicroFrontendResolver(microFrontendLister microFrontendLister) *microFrontendResolver {
	return &microFrontendResolver{
		microFrontendLister:    microFrontendLister,
		microFrontendConverter: newMicroFrontendConverter(),
	}
}

func (r *microFrontendResolver) MicroFrontendsQuery(ctx context.Context, namespace string) ([]gqlschema.MicroFrontend, error) {
	items, err := r.microFrontendLister.List(namespace)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.MicroFrontends))
		return nil, gqlerror.New(err, pretty.MicroFrontends)
	}

	mfs, err := r.microFrontendConverter.ToGQLs(items)
	if err != nil {
		return nil, err
	}
	return mfs, nil
}
