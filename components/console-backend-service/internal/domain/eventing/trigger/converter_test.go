package trigger

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestTriggerConverter_ToGQL(t *testing.T) {
	converter := NewTriggerConverter()
	rawURL := "www.test.com"
	url, _ := apis.ParseURL(rawURL)

	for testName, testData := range map[string]struct {
		toConvert  *v1alpha1.Trigger
		expected   *gqlschema.Trigger
		errMatcher types.GomegaMatcher
	}{
		"All properties with subscriber ref are given": {
			toConvert: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "TestName",
					Namespace: "TestNamespace",
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Filter: &v1alpha1.TriggerFilter{
						Attributes: &v1alpha1.TriggerFilterAttributes{
							"test1": "alpha", "test2": "beta",
						},
					},
					Subscriber: duckv1.Destination{
						Ref: &duckv1.KReference{
							Kind:       "TestKind",
							Namespace:  "TestNamespace",
							Name:       "TestName",
							APIVersion: "TestAPIVersion",
						},
					},
				},
				Status: v1alpha1.TriggerStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							apis.Condition{
								Status:  corev1.ConditionTrue,
								Message: "OK",
							},
						},
					},
				},
			},
			expected: &gqlschema.Trigger{
				Name:      "TestName",
				Namespace: "TestNamespace",
				Broker:    "default",
				FilterAttributes: gqlschema.JSON{
					"test1": "alpha", "test2": "beta",
				},
				Subscriber: gqlschema.Subscriber{
					Ref: &gqlschema.SubscriberRef{
						APIVersion: "TestAPIVersion",
						Kind:       "TestKind",
						Name:       "TestName",
						Namespace:  "TestNamespace",
					},
				},
				Status: gqlschema.TriggerStatus{
					Status: gqlschema.TriggerStatusTypeReady,
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"All properties with subscriber uri and error in status are given": {
			toConvert: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "TestName",
					Namespace: "TestNamespace",
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Filter: &v1alpha1.TriggerFilter{
						Attributes: &v1alpha1.TriggerFilterAttributes{
							"test1": "alpha", "test2": "beta",
						},
					},
					Subscriber: duckv1.Destination{
						URI: url,
					},
				},
				Status: v1alpha1.TriggerStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							apis.Condition{
								Status:  corev1.ConditionFalse,
								Message: "test error",
							},
							apis.Condition{
								Status:  corev1.ConditionFalse,
								Message: "test error",
							},
							apis.Condition{
								Status:  corev1.ConditionTrue,
								Message: "OK",
							},
						},
					},
				},
			},
			expected: &gqlschema.Trigger{
				Name:      "TestName",
				Namespace: "TestNamespace",
				Broker:    "default",
				FilterAttributes: gqlschema.JSON{
					"test1": "alpha", "test2": "beta",
				},
				Subscriber: gqlschema.Subscriber{
					URI: &rawURL,
				},
				Status: gqlschema.TriggerStatus{
					Status: gqlschema.TriggerStatusTypeFailed,
					Reason: []string{"test error", "test error"},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"All properties with different statuses": {
			toConvert: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "TestName",
					Namespace: "TestNamespace",
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Subscriber: duckv1.Destination{
						URI: url,
					},
				},
				Status: v1alpha1.TriggerStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							apis.Condition{
								Status:  corev1.ConditionUnknown,
								Message: "test unknown",
							},
							apis.Condition{
								Status:  corev1.ConditionFalse,
								Message: "test error",
							},
							apis.Condition{
								Status:  corev1.ConditionTrue,
								Message: "OK",
							},
						},
					},
				},
			},
			expected: &gqlschema.Trigger{
				Name:      "TestName",
				Namespace: "TestNamespace",
				Broker:    "default",
				Subscriber: gqlschema.Subscriber{
					URI: &rawURL,
				},
				Status: gqlschema.TriggerStatus{
					Status: gqlschema.TriggerStatusTypeFailed,
					Reason: []string{"test unknown", "test error"},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"All properties with status unknown": {
			toConvert: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "TestName",
					Namespace: "TestNamespace",
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Subscriber: duckv1.Destination{
						URI: url,
					},
				},
				Status: v1alpha1.TriggerStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							apis.Condition{
								Status:  corev1.ConditionUnknown,
								Message: "test unknown",
							},
							apis.Condition{
								Status:  corev1.ConditionTrue,
								Message: "OK",
							},
						},
					},
				},
			},
			expected: &gqlschema.Trigger{
				Name:      "TestName",
				Namespace: "TestNamespace",
				Broker:    "default",
				Subscriber: gqlschema.Subscriber{
					URI: &rawURL,
				},
				Status: gqlschema.TriggerStatus{
					Status: gqlschema.TriggerStatusTypeUnknown,
					Reason: []string{"test unknown"},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"Empty": {
			toConvert:  new(v1alpha1.Trigger),
			expected:   nil,
			errMatcher: gomega.HaveOccurred(),
		},
		"Nil": {
			toConvert:  nil,
			expected:   nil,
			errMatcher: gomega.HaveOccurred(),
		},
		"Empty Subscriber": {
			toConvert: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "TestName",
					Namespace: "TestNamespace",
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Filter: &v1alpha1.TriggerFilter{
						Attributes: &v1alpha1.TriggerFilterAttributes{
							"test1": "alpha", "test2": "beta",
						},
					},
				},
			},
			expected:   nil,
			errMatcher: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			//when
			converted, err := converter.ToGQL(testData.toConvert)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(converted).To(gomega.Equal(testData.expected))
		})
	}
}

func TestTriggerConverter_ToGQLs(t *testing.T) {
	converter := NewTriggerConverter()
	rawURL := "www.test.com"
	url, _ := apis.ParseURL(rawURL)

	for testName, testData := range map[string]struct {
		toConvert  []*v1alpha1.Trigger
		expected   []gqlschema.Trigger
		errMatcher types.GomegaMatcher
	}{
		"All properties are given": {
			toConvert: []*v1alpha1.Trigger{
				{
					TypeMeta: v1.TypeMeta{
						Kind:       "Trigger",
						APIVersion: "eventing.knative.dev/v1alpha1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "TestName",
						Namespace: "TestNamespace",
					},
					Spec: v1alpha1.TriggerSpec{
						Broker: "default",
						Filter: &v1alpha1.TriggerFilter{
							Attributes: &v1alpha1.TriggerFilterAttributes{
								"test1": "alpha", "test2": "beta",
							},
						},
						Subscriber: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "TestKind",
								Namespace:  "TestNamespace",
								Name:       "TestName",
								APIVersion: "TestAPIVersion",
							},
						},
					},
					Status: v1alpha1.TriggerStatus{
						Status: duckv1.Status{
							Conditions: duckv1.Conditions{
								apis.Condition{
									Status:  corev1.ConditionTrue,
									Message: "OK",
								},
							},
						},
					},
				},
				{
					TypeMeta: v1.TypeMeta{
						Kind:       "Trigger",
						APIVersion: "eventing.knative.dev/v1alpha1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "TestName",
						Namespace: "TestNamespace",
					},
					Spec: v1alpha1.TriggerSpec{
						Broker: "default",
						Filter: &v1alpha1.TriggerFilter{
							Attributes: &v1alpha1.TriggerFilterAttributes{
								"test1": "alpha", "test2": "beta",
							},
						},
						Subscriber: duckv1.Destination{
							URI: url,
						},
					},
					Status: v1alpha1.TriggerStatus{
						Status: duckv1.Status{
							Conditions: duckv1.Conditions{
								apis.Condition{
									Status:  corev1.ConditionFalse,
									Message: "test error",
								},
								apis.Condition{
									Status:  corev1.ConditionUnknown,
									Message: "test error",
								},
								apis.Condition{
									Status:  corev1.ConditionTrue,
									Message: "OK",
								},
							},
						},
					},
				},
			},
			expected: []gqlschema.Trigger{
				{
					Name:      "TestName",
					Namespace: "TestNamespace",
					Broker:    "default",
					FilterAttributes: gqlschema.JSON{
						"test1": "alpha", "test2": "beta",
					},
					Subscriber: gqlschema.Subscriber{
						Ref: &gqlschema.SubscriberRef{
							APIVersion: "TestAPIVersion",
							Kind:       "TestKind",
							Name:       "TestName",
							Namespace:  "TestNamespace",
						},
					},
					Status: gqlschema.TriggerStatus{
						Status: gqlschema.TriggerStatusTypeReady,
					},
				},
				{
					Name:      "TestName",
					Namespace: "TestNamespace",
					Broker:    "default",
					FilterAttributes: gqlschema.JSON{
						"test1": "alpha", "test2": "beta",
					},
					Subscriber: gqlschema.Subscriber{
						URI: &rawURL,
					},
					Status: gqlschema.TriggerStatus{
						Status: gqlschema.TriggerStatusTypeFailed,
						Reason: []string{"test error", "test error"},
					},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"Empty": {
			toConvert:  []*v1alpha1.Trigger{},
			expected:   []gqlschema.Trigger{},
			errMatcher: gomega.BeNil(),
		},
		"Nil": {
			toConvert:  nil,
			expected:   nil,
			errMatcher: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			//when
			converted, err := converter.ToGQLs(testData.toConvert)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(converted).To(gomega.Equal(testData.expected))
		})
	}
}

func TestTriggerConverter_ToTrigger(t *testing.T) {
	converter := NewTriggerConverter()
	rawURL := "www.test.com"
	triggerName := "TestName"
	url, _ := apis.ParseURL(rawURL)

	for testName, testData := range map[string]struct {
		toConvert  gqlschema.TriggerCreateInput
		ownerRef   []gqlschema.OwnerReference
		expected   *v1alpha1.Trigger
		errMatcher types.GomegaMatcher
	}{
		"All properties with subscriber ref are given": {
			toConvert: gqlschema.TriggerCreateInput{
				Name:      &triggerName,
				Namespace: "TestNamespace",
				Broker:    "default",
				FilterAttributes: &gqlschema.JSON{
					"test1": "alpha", "test2": "beta",
				},
				Subscriber: gqlschema.SubscriberInput{
					Ref: &gqlschema.SubscriberRefInput{
						Kind:       "TestKind",
						Namespace:  "TestNamespace",
						Name:       "TestName",
						APIVersion: "TestAPIVersion",
					},
				},
			},
			ownerRef: []gqlschema.OwnerReference{},
			expected: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:            "TestName",
					Namespace:       "TestNamespace",
					OwnerReferences: []v1.OwnerReference{},
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Filter: &v1alpha1.TriggerFilter{
						Attributes: &v1alpha1.TriggerFilterAttributes{
							"test1": "alpha", "test2": "beta",
						},
					},
					Subscriber: duckv1.Destination{
						Ref: &duckv1.KReference{
							Kind:       "TestKind",
							Namespace:  "TestNamespace",
							Name:       "TestName",
							APIVersion: "TestAPIVersion",
						},
					},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"All properties with subscriber uri are given": {
			toConvert: gqlschema.TriggerCreateInput{
				Name:      &triggerName,
				Namespace: "TestNamespace",
				Broker:    "default",
				Subscriber: gqlschema.SubscriberInput{
					URI: &rawURL,
				},
				FilterAttributes: &gqlschema.JSON{
					"test1": "alpha", "test2": "beta",
				},
			},
			ownerRef: []gqlschema.OwnerReference{},
			expected: &v1alpha1.Trigger{
				TypeMeta: v1.TypeMeta{
					Kind:       "Trigger",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:            "TestName",
					Namespace:       "TestNamespace",
					OwnerReferences: []v1.OwnerReference{},
				},
				Spec: v1alpha1.TriggerSpec{
					Broker: "default",
					Filter: &v1alpha1.TriggerFilter{
						Attributes: &v1alpha1.TriggerFilterAttributes{
							"test1": "alpha", "test2": "beta",
						},
					},
					Subscriber: duckv1.Destination{
						URI: url,
					},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"Empty": {
			toConvert:  gqlschema.TriggerCreateInput{},
			expected:   nil,
			errMatcher: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			//when
			converted, err := converter.ToTrigger(testData.toConvert, testData.ownerRef)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(converted).To(gomega.Equal(testData.expected))
		})
	}
}

func TestTriggerConverter_ToTriggers(t *testing.T) {
	converter := NewTriggerConverter()
	rawURL := "www.test.com"
	triggerName := "TestName"
	url, _ := apis.ParseURL(rawURL)

	for testName, testData := range map[string]struct {
		toConvert  []gqlschema.TriggerCreateInput
		ownerRef   []gqlschema.OwnerReference
		expected   []*v1alpha1.Trigger
		errMatcher types.GomegaMatcher
	}{
		"All properties are given": {
			toConvert: []gqlschema.TriggerCreateInput{
				{
					Name:      &triggerName,
					Namespace: "TestNamespace",
					Broker:    "default",
					FilterAttributes: &gqlschema.JSON{
						"test1": "alpha", "test2": "beta",
					},
					Subscriber: gqlschema.SubscriberInput{
						Ref: &gqlschema.SubscriberRefInput{
							Kind:       "TestKind",
							Namespace:  "TestNamespace",
							Name:       "TestName",
							APIVersion: "TestAPIVersion",
						},
					},
				},
				{
					Name:      &triggerName,
					Namespace: "TestNamespace",
					Broker:    "default",
					Subscriber: gqlschema.SubscriberInput{
						URI: &rawURL,
					},
					FilterAttributes: &gqlschema.JSON{
						"test1": "alpha", "test2": "beta",
					},
				},
			},
			ownerRef: []gqlschema.OwnerReference{
				{
					APIVersion:         "TestAPIVersion",
					Kind:               "TestKind",
					Name:               "TestName",
					BlockOwnerDeletion: new(bool),
					Controller:         new(bool),
				},
				{
					APIVersion:         "TestAPIVersionNext",
					Kind:               "TestKindNext",
					Name:               "TestNameNext",
					BlockOwnerDeletion: new(bool),
					Controller:         new(bool),
				},
			},
			expected: []*v1alpha1.Trigger{
				{
					TypeMeta: v1.TypeMeta{
						Kind:       "Trigger",
						APIVersion: "eventing.knative.dev/v1alpha1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "TestName",
						Namespace: "TestNamespace",
						OwnerReferences: []v1.OwnerReference{
							{
								APIVersion:         "TestAPIVersion",
								Kind:               "TestKind",
								Name:               "TestName",
								BlockOwnerDeletion: new(bool),
								Controller:         new(bool),
							},
							{
								APIVersion:         "TestAPIVersionNext",
								Kind:               "TestKindNext",
								Name:               "TestNameNext",
								BlockOwnerDeletion: new(bool),
								Controller:         new(bool),
							},
						},
					},
					Spec: v1alpha1.TriggerSpec{
						Broker: "default",
						Filter: &v1alpha1.TriggerFilter{
							Attributes: &v1alpha1.TriggerFilterAttributes{
								"test1": "alpha", "test2": "beta",
							},
						},
						Subscriber: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "TestKind",
								Namespace:  "TestNamespace",
								Name:       "TestName",
								APIVersion: "TestAPIVersion",
							},
						},
					},
				},
				{
					TypeMeta: v1.TypeMeta{
						Kind:       "Trigger",
						APIVersion: "eventing.knative.dev/v1alpha1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "TestName",
						Namespace: "TestNamespace",
						OwnerReferences: []v1.OwnerReference{
							{
								APIVersion:         "TestAPIVersion",
								Kind:               "TestKind",
								Name:               "TestName",
								BlockOwnerDeletion: new(bool),
								Controller:         new(bool),
							},
							{
								APIVersion:         "TestAPIVersionNext",
								Kind:               "TestKindNext",
								Name:               "TestNameNext",
								BlockOwnerDeletion: new(bool),
								Controller:         new(bool),
							},
						},
					},
					Spec: v1alpha1.TriggerSpec{
						Broker: "default",
						Filter: &v1alpha1.TriggerFilter{
							Attributes: &v1alpha1.TriggerFilterAttributes{
								"test1": "alpha", "test2": "beta",
							},
						},
						Subscriber: duckv1.Destination{
							URI: url,
						},
					},
				},
			},
			errMatcher: gomega.BeNil(),
		},
		"Empty": {
			toConvert:  []gqlschema.TriggerCreateInput{},
			expected:   []*v1alpha1.Trigger{},
			errMatcher: gomega.BeNil(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			//when
			converted, err := converter.ToTriggers(testData.toConvert, testData.ownerRef)

			//then
			g.Expect(err).To(testData.errMatcher)
			g.Expect(converted).To(gomega.ContainElements(testData.expected))
		})
	}
}
