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

type microFrontendResolver struct {
	service   *resource.Service
	converter MicroFrontendConverter
}

func newMicroFrontendResolver(sf *resource.ServiceFactory) *microFrontendResolver {
	return &microFrontendResolver{
		service:   sf.ForResource(v1alpha1.SchemeGroupVersion.WithResource("microfrontends")),
		converter: NewMicroFrontendConverter(),
	}
}

func (r *microFrontendResolver) MicroFrontendsQuery(ctx context.Context, namespace string) ([]*model.MicroFrontend, error) {
	list := MicroFrontendList{}
	err := r.service.ListInNamespace(namespace, &list)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.MicroFrontends))
		return nil, gqlerror.New(err, pretty.MicroFrontends)
	}

	mfs, err := r.converter.ToGQLs(list)
	if err != nil {
		return nil, err
	}
	return mfs, nil
}
