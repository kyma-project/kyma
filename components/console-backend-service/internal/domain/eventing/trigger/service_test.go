package trigger

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const timeout = time.Second * 3

func TestTriggerService_List(t *testing.T) {
	url := "www.test.com"
	trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
	trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
	trigger3 := fixTriggerWithRef("a3", "a", "refA2", "refA")
	trigger4 := fixTriggerWithRef("b1", "b", "refA1", "refA")
	trigger5 := fixTriggerWithUri("a4", "a", url)
	trigger6 := fixTriggerWithUri("a5", "a", url)

	for testName, testData := range map[string]struct {
		namespace       string
		subscriberInput *gqlschema.SubscriberInput
		errMatcher      types.GomegaMatcher
		containElements []interface{}
	}{
		"Success with given namespace only": {
			namespace:       "a",
			subscriberInput: nil,
			errMatcher:      gomega.BeNil(),
			containElements: []interface{}{
				trigger1, trigger2, trigger3, trigger5, trigger6,
			},
		},
		"Success with given namespace and ref": {
			namespace: "a",
			subscriberInput: &gqlschema.SubscriberInput{
				Ref: &gqlschema.SubscriberRefInput{
					Kind:       "TestKind",
					Namespace:  "refA",
					Name:       "refA1",
					APIVersion: "TestAPIVersion",
				},
			},
			errMatcher: gomega.BeNil(),
			containElements: []interface{}{
				trigger1, trigger2,
			},
		},
		"Success with given namespace and uri": {
			namespace: "a",
			subscriberInput: &gqlschema.SubscriberInput{
				URI: &url,
			},
			errMatcher: gomega.BeNil(),
			containElements: []interface{}{
				trigger5, trigger6,
			},
		},
		"Empty": {
			namespace:       "",
			subscriberInput: nil,
			errMatcher:      gomega.BeNil(),
			containElements: []interface{}{},
		},
		"With subscriberInput without namespace": {
			namespace: "",
			subscriberInput: &gqlschema.SubscriberInput{
				URI: &url,
			},
			errMatcher:      gomega.BeNil(),
			containElements: []interface{}{},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewGomegaWithT(t)
			service := fixTriggerService(t, trigger1, trigger2, trigger3, trigger4, trigger5, trigger6)

			//when
			list, err := service.List(testData.namespace, testData.subscriberInput)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(list).To(gomega.ContainElements(testData.containElements))
		})
	}
}

func TestTriggerService_Create(t *testing.T) {
	for testName, testData := range map[string]struct {
		trigger        *v1alpha1.Trigger
		errMatcher     types.GomegaMatcher
		triggerMatcher types.GomegaMatcher
	}{
		"Success": {
			trigger:        fixTriggerWithUri("TestName", "TestNamespace", "www.test.com"),
			errMatcher:     gomega.BeNil(),
			triggerMatcher: gomega.Not(gomega.BeNil()),
		},
		"Nil": {
			trigger:        nil,
			errMatcher:     gomega.HaveOccurred(),
			triggerMatcher: gomega.BeNil(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewGomegaWithT(t)
			service := fixTriggerService(t)

			//when
			created, err := service.Create(testData.trigger)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(created).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerService_CreateMany(t *testing.T) {
	url := "www.test.com"

	for testName, testData := range map[string]struct {
		triggers       []*v1alpha1.Trigger
		errMatcher     types.GomegaMatcher
		triggerMatcher types.GomegaMatcher
	}{
		"Success": {
			triggers: []*v1alpha1.Trigger{
				fixTriggerWithUri("a1", "a", url),
				fixTriggerWithUri("a2", "a", url),
				fixTriggerWithRef("a1", "b", "refA1", "refA"),
			},
			errMatcher:     gomega.BeNil(),
			triggerMatcher: gomega.HaveLen(3),
		},
		"Empty": {
			triggers:       []*v1alpha1.Trigger{},
			errMatcher:     gomega.BeNil(),
			triggerMatcher: gomega.HaveLen(0),
		},
		"Nil": {
			triggers:       nil,
			errMatcher:     gomega.BeNil(),
			triggerMatcher: gomega.HaveLen(0),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewGomegaWithT(t)
			service := fixTriggerService(t)

			//when
			created, err := service.CreateMany(testData.triggers)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(created).To(testData.triggerMatcher)
		})
	}
}

func TestTriggerService_Delete(t *testing.T) {
	url := "www.test.com"
	trigger1 := fixTriggerWithUri("a1", "a", url)
	trigger2 := fixTriggerWithUri("a2", "a", url)
	trigger3 := fixTriggerWithRef("a1", "b", "refA1", "refA")

	for testName, testData := range map[string]struct {
		trigger    gqlschema.TriggerMetadataInput
		errMatcher types.GomegaMatcher
	}{
		"Success": {
			trigger:    gqlschema.TriggerMetadataInput{Name: "a2", Namespace: "a"},
			errMatcher: gomega.BeNil(),
		},
		"Without namespace": {
			trigger:    gqlschema.TriggerMetadataInput{Name: "a2"},
			errMatcher: gomega.HaveOccurred(),
		},
		"Without name": {
			trigger:    gqlschema.TriggerMetadataInput{Namespace: "a"},
			errMatcher: gomega.HaveOccurred(),
		},
		"empty": {
			trigger:    gqlschema.TriggerMetadataInput{},
			errMatcher: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewGomegaWithT(t)
			service := fixTriggerService(t, trigger1, trigger2, trigger3)

			//when
			err := service.Delete(testData.trigger)

			//then
			g.Expect(err).To(testData.errMatcher)
		})
	}
}

func TestTriggerService_DeleteMany(t *testing.T) {
	url := "www.test.com"
	trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
	trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
	trigger3 := fixTriggerWithRef("a3", "a", "refA2", "refA")
	trigger4 := fixTriggerWithRef("b1", "b", "refA1", "refA")
	trigger5 := fixTriggerWithUri("a4", "a", url)
	trigger6 := fixTriggerWithUri("a5", "a", url)

	for testName, testData := range map[string]struct {
		triggers   []gqlschema.TriggerMetadataInput
		errMatcher types.GomegaMatcher
	}{
		"Success": {
			triggers: []gqlschema.TriggerMetadataInput{
				{Name: "a3", Namespace: "a"}, {Name: "a1", Namespace: "a"}, {Name: "a2", Namespace: "a"}, {Name: "b1", Namespace: "b"},
			},
			errMatcher: gomega.BeNil(),
		},
		"Empty": {
			triggers:   []gqlschema.TriggerMetadataInput{},
			errMatcher: gomega.BeNil(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewGomegaWithT(t)
			service := fixTriggerService(t, trigger1, trigger2, trigger3, trigger4, trigger5, trigger6)

			//when
			err := service.DeleteMany(testData.triggers)

			//then
			g.Expect(err).To(testData.errMatcher)
		})
	}
}

func TestTriggerService_SubscribeAndUnsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		//given
		trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
		trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
		service := fixTriggerService(t, trigger1, trigger2)

		//when
		extractor := extractor.TriggerUnstructuredExtractor{}
		listenerA := listener.NewTrigger(extractor, nil, nil, nil)

		service.Subscribe(listenerA)

		service.Unsubscribe(listenerA)
	})

	t.Run("Duplicated", func(t *testing.T) {
		//given
		trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
		trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
		service := fixTriggerService(t, trigger1, trigger2)

		//when
		extractor := extractor.TriggerUnstructuredExtractor{}
		listenerA := listener.NewTrigger(extractor, nil, nil, nil)

		service.Subscribe(listenerA)
		service.Subscribe(listenerA)

		service.Unsubscribe(listenerA)
	})

	t.Run("Multiple", func(t *testing.T) {
		//given
		//g := gomega.NewGomegaWithT(t)
		trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
		trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
		service := fixTriggerService(t, trigger1, trigger2)

		//when
		extractor := extractor.TriggerUnstructuredExtractor{}
		listenerA := listener.NewTrigger(extractor, nil, nil, nil)
		listenerB := listener.NewTrigger(extractor, nil, nil, nil)

		service.Subscribe(listenerA)
		service.Subscribe(listenerB)

		service.Unsubscribe(listenerA)
		service.Unsubscribe(listenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		//given
		//g := gomega.NewGomegaWithT(t)
		trigger1 := fixTriggerWithRef("a1", "a", "refA1", "refA")
		trigger2 := fixTriggerWithRef("a2", "a", "refA1", "refA")
		service := fixTriggerService(t, trigger1, trigger2)

		//when

		service.Subscribe(nil)
		service.Unsubscribe(nil)
	})
}

func fixTrigger(name, namespace string) *v1alpha1.Trigger {
	return &v1alpha1.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: "default",
			Filter: &v1alpha1.TriggerFilter{
				Attributes: &v1alpha1.TriggerFilterAttributes{
					"test1": "alpha", "test2": "beta",
				},
			},
		},
		Status: v1alpha1.TriggerStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					apis.Condition{
						Status: corev1.ConditionTrue,
						Reason: "OK",
					},
				},
			},
		},
	}
}

func fixTriggerWithRef(name, namespace, refName, refNamespace string) *v1alpha1.Trigger {
	trigger := fixTrigger(name, namespace)
	trigger.Spec.Subscriber.Ref = fixRef(refName, refNamespace)
	return trigger
}

func fixTriggerWithUri(name, namespace, url string) *v1alpha1.Trigger {
	trigger := fixTrigger(name, namespace)
	trigger.Spec.Subscriber.URI = fixUri(url)
	return trigger
}

func fixRef(name, namespace string) *duckv1.KReference {
	return &duckv1.KReference{
		Kind:       "TestKind",
		Namespace:  namespace,
		Name:       name,
		APIVersion: "TestAPIVersion",
	}
}

func fixUri(url string) *apis.URL {
	uri, _ := apis.ParseURL(url)
	return uri
}

func fixTriggerService(t *testing.T, objects ...runtime.Object) Service {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, objects...)
	require.NoError(t, err)

	extractor := extractor.TriggerUnstructuredExtractor{}
	service, err := NewService(serviceFactory, extractor)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, timeout, service.Informer)

	return service
}
