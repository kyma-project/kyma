package ui

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
)

//go:generate mockery -name=microfrontendSvc -output=automock -outpkg=automock -case=underscore
type microfrontendSvc interface {
	List(namespace string, pagingParams pager.PagingParams) ([]*v1alpha1.MicroFrontend, error)
}

//go:generate mockery -name=gqlMicrofrontendConverter -output=automock -outpkg=automock -case=underscore
type gqlMicrofrontendConverter interface {
	ToGQL(in *v1alpha1.MicroFrontend) (*gqlschema.Microfrontend, error)
	ToGQLs(in []*v1alpha1.MicroFrontend) ([]gqlschema.Microfrontend, error)
}

type microfrontendResolver struct {
	microfrontendSvc       microfrontendSvc
	microfrontendConverter gqlMicrofrontendConverter
}

func newMicrofrontendResolver(microfrontendSvc microfrontendSvc) *microfrontendResolver {
	return &microfrontendResolver{
		microfrontendSvc:       microfrontendSvc,
		microfrontendConverter: &microfrontendConverter{},
	}
}

func (r *microfrontendResolver) MicrofrontendsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Microfrontend, error) {
	var items []*v1alpha1.MicroFrontend
	var err error

	items, err = r.microfrontendSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

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
