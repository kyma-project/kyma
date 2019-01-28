package application

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/pretty"
	contentPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type eventActivationResolver struct {
	service          eventActivationLister
	converter        *eventActivationConverter
	contentRetriever shared.ContentRetriever
}

//go:generate mockery -name=eventActivationLister -output=automock -outpkg=automock -case=underscore
type eventActivationLister interface {
	List(environment string) ([]*v1alpha1.EventActivation, error)
}

func newEventActivationResolver(service eventActivationLister, contentRetriever shared.ContentRetriever) *eventActivationResolver {
	return &eventActivationResolver{
		service:          service,
		converter:        &eventActivationConverter{},
		contentRetriever: contentRetriever,
	}
}

func (r *eventActivationResolver) EventActivationsQuery(ctx context.Context, namespace string) ([]gqlschema.EventActivation, error) {
	items, err := r.service.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s in `%s` namespace", pretty.EventActivations, namespace))
		return nil, gqlerror.New(err, pretty.EventActivations, gqlerror.WithEnvironment(namespace))
	}

	return r.converter.ToGQLs(items), nil
}

func (r *eventActivationResolver) EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	if eventActivation == nil {
		glog.Errorf("EventActivation cannot be empty in order to resolve events field")
		return nil, gqlerror.NewInternal()
	}

	asyncApiSpec, err := r.contentRetriever.AsyncApiSpec().Find("service-class", eventActivation.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.EventActivationEvents, pretty.EventActivation, eventActivation.Name))
		return nil, gqlerror.New(err, pretty.EventActivationEvents, gqlerror.WithName(eventActivation.Name))
	}

	if asyncApiSpec == nil {
		return []gqlschema.EventActivationEvent{}, nil
	}

	if asyncApiSpec.Data.AsyncAPI != "1.0.0" {
		details := fmt.Sprintf("not supported version `%s` of %s", asyncApiSpec.Data.AsyncAPI, contentPretty.AsyncApiSpec)
		glog.Error(details)
		return nil, gqlerror.NewInternal(gqlerror.WithDetails(details))
	}

	return r.converter.ToGQLEvents(asyncApiSpec), nil
}
