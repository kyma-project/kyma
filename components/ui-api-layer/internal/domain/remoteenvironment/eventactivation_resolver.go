package remoteenvironment

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	contentPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type eventActivationResolver struct {
	service            eventActivationLister
	converter          *eventActivationConverter
	asyncApiSpecGetter AsyncApiSpecGetter
}

func newEventActivationResolver(service eventActivationLister, asyncApiSpecGetter AsyncApiSpecGetter) *eventActivationResolver {
	return &eventActivationResolver{
		service:            service,
		converter:          &eventActivationConverter{},
		asyncApiSpecGetter: asyncApiSpecGetter,
	}
}

func (r *eventActivationResolver) EventActivationsQuery(ctx context.Context, environment string) ([]gqlschema.EventActivation, error) {
	items, err := r.service.List(environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s in `%s` environment", pretty.EventActivations, environment))
		return nil, gqlerror.New(err, pretty.EventActivations, gqlerror.WithEnvironment(environment))
	}

	return r.converter.ToGQLs(items), nil
}

func (r *eventActivationResolver) EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	if eventActivation == nil {
		glog.Errorf("EventActivation cannot be empty in order to resolve events field")
		return nil, gqlerror.NewInternal()
	}

	asyncApiSpec, err := r.asyncApiSpecGetter.Find("service-class", eventActivation.Name)
	if err != nil {
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
