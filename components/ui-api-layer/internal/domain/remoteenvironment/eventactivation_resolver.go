package remoteenvironment

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
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
		glog.Error(errors.Wrapf(err, "while listing eventActivations in `%s` environment", environment))
		return nil, err
	}

	return r.converter.ToGQLs(items), nil
}

func (r *eventActivationResolver) EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	if eventActivation == nil {
		glog.Errorf("EventActivation cannot be empty in order to resolve events field")
		return nil, r.eventsError()
	}

	asyncApiSpec, err := r.asyncApiSpecGetter.Find("service-class", eventActivation.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering events for EventActivation %s", eventActivation.Name))
		return nil, r.eventsError()
	}

	if asyncApiSpec == nil {
		return []gqlschema.EventActivationEvent{}, nil
	}

	if asyncApiSpec.Data.AsyncAPI != "1.0.0" {
		glog.Errorf("not supported version `%s` of asyncApiSpec", asyncApiSpec.Data.AsyncAPI)
		return nil, r.eventsError()
	}

	return r.converter.ToGQLEvents(asyncApiSpec), nil
}

func (r *eventActivationResolver) genericError() error {
	return errors.New("Cannot get EventActivation")
}

func (r *eventActivationResolver) eventsError() error {
	return errors.New("Cannot list Events")
}
