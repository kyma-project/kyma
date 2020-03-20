package eventing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/trigger/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
)

func TestTriggerResolver_TriggersQuery(t *testing.T) {
	for testName, testData := range map[string]struct {
		namespace      string
		subscriber     *gqlschema.SubscriberInput
		triggerMatcher types.GomegaMatcher
		errorMatcher   types.GomegaMatcher

		//Mocks
		list       []*v1alpha1.Trigger
		listError  error
		toGQL      []gqlschema.Trigger
		toGQLError error
	}{
		"Success with subscriber": {
			namespace:      "test",
			subscriber:     &gqlschema.SubscriberInput{},
			list:           []*v1alpha1.Trigger{},
			listError:      nil,
			toGQL:          []gqlschema.Trigger{},
			toGQLError:     nil,
			triggerMatcher: gomega.Not(gomega.BeNil()),
			errorMatcher:   gomega.BeNil(),
		},
		"Success without subscriber": {
			namespace:      "test",
			subscriber:     nil,
			list:           []*v1alpha1.Trigger{},
			listError:      nil,
			toGQL:          []gqlschema.Trigger{},
			toGQLError:     nil,
			triggerMatcher: gomega.Not(gomega.BeNil()),
			errorMatcher:   gomega.BeNil(),
		},
		"Listing error": {
			namespace:      "test",
			subscriber:     nil,
			list:           nil,
			listError:      errors.New(""),
			toGQL:          []gqlschema.Trigger{},
			toGQLError:     nil,
			triggerMatcher: gomega.HaveLen(0),
			errorMatcher:   gomega.HaveOccurred(),
		},
		"Converting error": {
			namespace:      "test",
			subscriber:     nil,
			list:           []*v1alpha1.Trigger{},
			listError:      nil,
			toGQL:          []gqlschema.Trigger{},
			toGQLError:     errors.New(""),
			triggerMatcher: gomega.HaveLen(0),
			errorMatcher:   gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)
			ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
			cancel()
			service := &automock.Service{}
			converter := &automock.GQLConverter{}
			extractor := extractor.TriggerUnstructuredExtractor{}
			service.On(
				"List", testData.namespace, testData.subscriber,
			).Return(testData.list, testData.listError)
			converter.On(
				"ToGQLs", testData.list,
			).Return(testData.toGQL, testData.toGQLError)

			//when
			res := newTriggerResolver(service, converter, extractor, name.Generate)
			trigger, err := res.TriggersQuery(ctx, testData.namespace, testData.subscriber)

			//then
			g.Expect(err).To(testData.errorMatcher)
			g.Expect(trigger).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerResolver_CreateTrigger(t *testing.T) {
	triggerName := "TestName"
	for testName, testData := range map[string]struct {
		trigger        gqlschema.TriggerCreateInput
		ownerRef       []gqlschema.OwnerReference
		triggerMatcher types.GomegaMatcher
		errorMatcher   types.GomegaMatcher

		//Mocks
		toTrigger          *v1alpha1.Trigger
		toTriggerError     error
		createTrigger      *v1alpha1.Trigger
		createTriggerError error
		toGQL              *gqlschema.Trigger
		toGQLError         error
	}{
		"Success": {
			trigger: gqlschema.TriggerCreateInput{
				Name: &triggerName,
			},
			ownerRef:           []gqlschema.OwnerReference{},
			toTrigger:          &v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTrigger:      &v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQL:              &gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.Not(gomega.BeNil()),
			errorMatcher:       gomega.BeNil(),
		},
		"ToTrigger error": {
			trigger: gqlschema.TriggerCreateInput{
				Name: &triggerName,
			},
			ownerRef:           []gqlschema.OwnerReference{},
			toTrigger:          &v1alpha1.Trigger{},
			toTriggerError:     errors.New(""),
			createTrigger:      &v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQL:              &gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
		"List error": {
			trigger: gqlschema.TriggerCreateInput{
				Name: &triggerName,
			},
			ownerRef:           []gqlschema.OwnerReference{},
			toTrigger:          &v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTrigger:      &v1alpha1.Trigger{},
			createTriggerError: errors.New(""),
			toGQL:              &gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
		"ToGQL error": {
			trigger: gqlschema.TriggerCreateInput{
				Name: &triggerName,
			},
			ownerRef:           []gqlschema.OwnerReference{},
			toTrigger:          &v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTrigger:      &v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQL:              &gqlschema.Trigger{},
			toGQLError:         errors.New(""),
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)
			ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
			cancel()
			service := &automock.Service{}
			converter := &automock.GQLConverter{}
			extractor := extractor.TriggerUnstructuredExtractor{}
			converter.On(
				"ToTrigger", testData.trigger, testData.ownerRef,
			).Return(testData.toTrigger, testData.toTriggerError)
			service.On(
				"Create", testData.toTrigger,
			).Return(testData.createTrigger, testData.createTriggerError)
			converter.On(
				"ToGQL", testData.createTrigger,
			).Return(testData.toGQL, testData.toGQLError)

			//when
			res := newTriggerResolver(service, converter, extractor, name.Generate)
			trigger, err := res.CreateTrigger(ctx, testData.trigger, testData.ownerRef)

			//then
			g.Expect(err).To(testData.errorMatcher)
			g.Expect(trigger).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerResolver_CreateTriggers(t *testing.T) {
	for testName, testData := range map[string]struct {
		triggers       []gqlschema.TriggerCreateInput
		ownerRef       []gqlschema.OwnerReference
		triggerMatcher types.GomegaMatcher
		errorMatcher   types.GomegaMatcher

		//Mocks
		toTriggers         []*v1alpha1.Trigger
		toTriggerError     error
		createTriggers     []*v1alpha1.Trigger
		createTriggerError error
		toGQLs             []gqlschema.Trigger
		toGQLError         error
	}{
		"Success": {
			triggers:           []gqlschema.TriggerCreateInput{},
			ownerRef:           []gqlschema.OwnerReference{},
			toTriggers:         []*v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTriggers:     []*v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQLs:             []gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.Not(gomega.BeNil()),
			errorMatcher:       gomega.BeNil(),
		},
		"ToTriggers error": {
			triggers:           []gqlschema.TriggerCreateInput{},
			ownerRef:           []gqlschema.OwnerReference{},
			toTriggers:         []*v1alpha1.Trigger{},
			toTriggerError:     errors.New(""),
			createTriggers:     []*v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQLs:             []gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
		"CreateMany error": {
			triggers:           []gqlschema.TriggerCreateInput{},
			ownerRef:           []gqlschema.OwnerReference{},
			toTriggers:         []*v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTriggers:     []*v1alpha1.Trigger{},
			createTriggerError: errors.New(""),
			toGQLs:             []gqlschema.Trigger{},
			toGQLError:         nil,
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
		"ToGQLs error": {
			triggers:           []gqlschema.TriggerCreateInput{},
			ownerRef:           []gqlschema.OwnerReference{},
			toTriggers:         []*v1alpha1.Trigger{},
			toTriggerError:     nil,
			createTriggers:     []*v1alpha1.Trigger{},
			createTriggerError: nil,
			toGQLs:             []gqlschema.Trigger{},
			toGQLError:         errors.New(""),
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)
			ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
			cancel()
			service := &automock.Service{}
			converter := &automock.GQLConverter{}
			extractor := extractor.TriggerUnstructuredExtractor{}
			converter.On(
				"ToTriggers", testData.triggers, testData.ownerRef,
			).Return(testData.toTriggers, testData.toTriggerError)
			service.On(
				"CreateMany", testData.toTriggers,
			).Return(testData.createTriggers, testData.createTriggerError)
			converter.On(
				"ToGQLs", testData.createTriggers,
			).Return(testData.toGQLs, testData.toGQLError)

			//when
			res := newTriggerResolver(service, converter, extractor, name.Generate)
			trigger, err := res.CreateManyTriggers(ctx, testData.triggers, testData.ownerRef)

			//then
			g.Expect(err).To(testData.errorMatcher)
			g.Expect(trigger).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerResolver_DeleteTrigger(t *testing.T) {
	for testName, testData := range map[string]struct {
		trigger        gqlschema.TriggerMetadataInput
		triggerMatcher types.GomegaMatcher
		errorMatcher   types.GomegaMatcher

		//Mocks
		deleteTriggerError error
	}{
		"Success": {
			trigger:            gqlschema.TriggerMetadataInput{Name: "a", Namespace: "a"},
			deleteTriggerError: nil,
			triggerMatcher:     gomega.BeEquivalentTo(&gqlschema.TriggerMetadataInput{Name: "a", Namespace: "a"}),
			errorMatcher:       gomega.BeNil(),
		},
		"Error": {
			trigger:            gqlschema.TriggerMetadataInput{},
			deleteTriggerError: errors.New(""),
			triggerMatcher:     gomega.BeNil(),
			errorMatcher:       gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)
			ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
			cancel()
			service := &automock.Service{}
			converter := &automock.GQLConverter{}
			extractor := extractor.TriggerUnstructuredExtractor{}
			service.On(
				"Delete", testData.trigger,
			).Return(testData.deleteTriggerError)

			//when
			res := newTriggerResolver(service, converter, extractor, name.Generate)
			trigger, err := res.DeleteTrigger(ctx, testData.trigger)

			//then
			g.Expect(err).To(testData.errorMatcher)
			g.Expect(trigger).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerResolver_DeleteManyTriggers(t *testing.T) {
	for testName, testData := range map[string]struct {
		triggers       []gqlschema.TriggerMetadataInput
		triggerMatcher types.GomegaMatcher
		errorMatcher   types.GomegaMatcher

		//Mocks
		deleteTriggerError error
	}{
		"Success": {
			triggers: []gqlschema.TriggerMetadataInput{
				{Name: "a1", Namespace: "a"}, {Name: "a2", Namespace: "a"},
			},
			deleteTriggerError: nil,
			triggerMatcher:     gomega.HaveLen(2),
			errorMatcher:       gomega.BeNil(),
		},
		"Error": {
			triggers: []gqlschema.TriggerMetadataInput{
				{Name: "a1", Namespace: "a"}, {Name: "a2", Namespace: "a"},
			},
			deleteTriggerError: errors.New(""),
			triggerMatcher:     gomega.HaveLen(0),
			errorMatcher:       gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)
			ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
			cancel()
			service := &automock.Service{}
			converter := &automock.GQLConverter{}
			extractor := extractor.TriggerUnstructuredExtractor{}
			service.On(
				"Delete", mock.Anything,
			).Return(testData.deleteTriggerError)

			//when
			res := newTriggerResolver(service, converter, extractor, name.Generate)
			trigger, err := res.DeleteManyTriggers(ctx, testData.triggers)

			//then
			g.Expect(err).To(testData.errorMatcher)
			g.Expect(trigger).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerResolver_TriggerEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//given
		g := gomega.NewWithT(t)
		ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
		cancel()
		service := &automock.Service{}
		converter := &automock.GQLConverter{}
		extractor := extractor.TriggerUnstructuredExtractor{}
		service.On("Subscribe", mock.Anything).Once()
		service.On("Unsubscribe", mock.Anything).Once()

		//when
		res := newTriggerResolver(service, converter, extractor, name.Generate)
		_, err := res.TriggerEventSubscription(ctx, "test", nil)

		//then
		g.Expect(err).To(gomega.BeNil())
		service.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		//given
		g := gomega.NewWithT(t)
		ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
		cancel()
		service := &automock.Service{}
		converter := &automock.GQLConverter{}
		extractor := extractor.TriggerUnstructuredExtractor{}
		service.On("Subscribe", mock.Anything).Once()
		service.On("Unsubscribe", mock.Anything).Once()

		//when
		res := newTriggerResolver(service, converter, extractor, name.Generate)
		channel, err := res.TriggerEventSubscription(ctx, "test", nil)
		<-channel

		//then
		g.Expect(err).To(gomega.BeNil())
		service.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
