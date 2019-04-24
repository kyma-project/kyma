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

//go:generate mockery -name=microfrontendLister -output=automock -outpkg=automock -case=underscore
type microfrontendLister interface {
	List(namespace string) ([]*v1alpha1.MicroFrontend, error)
}

//go:generate mockery -name=gqlMicrofrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlMicrofrontendConverter interface {
	ToGQL(in *v1alpha1.MicroFrontend) (*gqlschema.Microfrontend, error)
	ToGQLs(in []*v1alpha1.MicroFrontend) ([]gqlschema.Microfrontend, error)
}

type microfrontendResolver struct {
	microfrontendLister    microfrontendLister
	microfrontendConverter gqlMicrofrontendConverter
}

func newMicrofrontendResolver(microfrontendLister microfrontendLister) *microfrontendResolver {
	return &microfrontendResolver{
		microfrontendLister:    microfrontendLister,
		microfrontendConverter: &microfrontendConverter{},
	}
}

func (r *microfrontendResolver) MicrofrontendsQuery(ctx context.Context, namespace string) ([]gqlschema.Microfrontend, error) {
	var items []*v1alpha1.MicroFrontend
	var err error

	items, err = r.microfrontendLister.List(namespace)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Microfrontends))
		return nil, gqlerror.New(err, pretty.Microfrontends)
	}

	mfs, err := r.microfrontendConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Microfrontends))
		return nil, gqlerror.New(err, pretty.Microfrontends)
	}

	return mfs, nil
}
