package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate mockery -name=resourceSvc -output=automock -outpkg=automock -case=underscore
type resourceSvc interface {
	Create(namespace string, resource types.Resource) (*types.Resource, error)
}

//go:generate mockery -name=gqlResourceConverter -output=automock -outpkg=automock -case=underscore
type gqlResourceConverter interface {
	GQLJSONToResource(in gqlschema.JSON) (types.Resource, error)
	BodyToGQLJSON(in []byte) (gqlschema.JSON, error)
}

type resourceResolver struct {
	resourceSvc
	gqlResourceConverter
}

func newResourceResolver(resourceSvc resourceSvc) *resourceResolver {
	return &resourceResolver{
		resourceSvc:          resourceSvc,
		gqlResourceConverter: &resourceConverter{},
	}
}

func (r *resourceResolver) CreateResourceMutation(ctx context.Context, namespace string, resource gqlschema.JSON) (gqlschema.JSON, error) {
	converted, err := r.gqlResourceConverter.GQLJSONToResource(resource)
	if err != nil {
		return nil, gqlerror.New(err, pretty.Resource)
	}

	created, err := r.resourceSvc.Create(namespace, converted)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s %s `%s`", pretty.Resource, converted.Kind, converted.Name))
		return nil, gqlerror.New(err, pretty.Resource)
	}

	body, err := r.gqlResourceConverter.BodyToGQLJSON(created.Body)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s body from namespace %s", pretty.Resource, namespace))
		return nil, gqlerror.New(err, pretty.Pod, gqlerror.WithName(converted.Name), gqlerror.WithNamespace(namespace))
	}

	return body, nil
}
