package eventing

import (
	"context"

	"github.com/golang/glog"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/trigger"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/pkg/errors"
)

type triggerResolver struct {
	service   trigger.Service
	converter trigger.GQLConverter
	extractor extractor.TriggerUnstructuredExtractor

	generateName func() string
}

func newTriggerResolver(svc trigger.Service, converter trigger.GQLConverter, extractor extractor.TriggerUnstructuredExtractor, nameGenerator func() string) *triggerResolver {
	return &triggerResolver{
		service:      svc,
		converter:    converter,
		extractor:    extractor,
		generateName: nameGenerator,
	}
}

func (r *triggerResolver) TriggersQuery(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) ([]gqlschema.Trigger, error) {
	triggers, err := r.service.List(namespace, subscriber)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s in namespace `%s`", pretty.Triggers, namespace))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	out, err := r.converter.ToGQLs(triggers)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Triggers))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	return out, nil
}

func (r *triggerResolver) CreateTrigger(ctx context.Context, trigger gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*gqlschema.Trigger, error) {
	trigger = r.checkTriggerName(trigger)

	converted, err := r.converter.ToTrigger(trigger, ownerRef)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting input %s `%s`", pretty.Trigger, *trigger.Name))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	created, err := r.service.Create(converted)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.Trigger, *trigger.Name))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	out, err := r.converter.ToGQL(created)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting output %s", pretty.Triggers))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	return out, nil
}

func (r *triggerResolver) CreateManyTriggers(ctx context.Context, triggers []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]gqlschema.Trigger, error) {
	triggers = r.checkTriggersNames(triggers)

	converted, err := r.converter.ToTriggers(triggers, ownerRef)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting input %s", pretty.Triggers))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	created, err := r.service.CreateMany(converted)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s", pretty.Triggers))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	out, err := r.converter.ToGQLs(created)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting output %s", pretty.Triggers))
		return nil, gqlerror.New(err, pretty.Triggers)
	}

	return out, nil
}

func (r *triggerResolver) DeleteTrigger(ctx context.Context, trigger gqlschema.TriggerMetadataInput) (*gqlschema.TriggerMetadata, error) {
	err := r.service.Delete(trigger)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.Trigger, trigger.Name))
		return nil, gqlerror.New(err, pretty.Triggers)
	}
	return &gqlschema.TriggerMetadata{
		Name:      trigger.Name,
		Namespace: trigger.Namespace,
	}, nil
}

func (r *triggerResolver) DeleteManyTriggers(ctx context.Context, triggers []gqlschema.TriggerMetadataInput) ([]gqlschema.TriggerMetadata, error) {
	deletedTriggers := make([]gqlschema.TriggerMetadata, 0)
	for _, trigger := range triggers {
		_, err := r.DeleteTrigger(ctx, trigger)
		if err != nil {
			return deletedTriggers, err
		}
		deletedTriggers = append(deletedTriggers, gqlschema.TriggerMetadata{
			Name:      trigger.Name,
			Namespace: trigger.Namespace,
		})
	}

	return deletedTriggers, nil
}

func (r *triggerResolver) TriggerEventSubscription(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) (<-chan gqlschema.TriggerEvent, error) {
	channel := make(chan gqlschema.TriggerEvent, 1)
	filter := func(entity *v1alpha1.Trigger) bool {
		isSubCorrect := true
		if entity == nil {
			return false
		}
		if subscriber != nil {
			isSubCorrect = r.service.CompareSubscribers(subscriber, entity)
		}
		return entity.Namespace == namespace && isSubCorrect
	}

	listener := listener.NewTrigger(r.extractor, channel, filter, r.converter)

	r.service.Subscribe(listener)
	go func() {
		defer close(channel)
		defer r.service.Unsubscribe(listener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *triggerResolver) checkTriggerName(trigger gqlschema.TriggerCreateInput) gqlschema.TriggerCreateInput {
	if trigger.Name == nil || *trigger.Name == "" {
		name := r.generateName()
		trigger.Name = &name
	}
	return trigger
}

func (r *triggerResolver) checkTriggersNames(triggers []gqlschema.TriggerCreateInput) []gqlschema.TriggerCreateInput {
	for i, trigger := range triggers {
		triggers[i] = r.checkTriggerName(trigger)
	}
	return triggers
}
