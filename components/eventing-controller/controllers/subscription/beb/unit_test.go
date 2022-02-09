package beb

import (
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_isInDeletion(t *testing.T) {
	var tests = []struct {
		name              string
		givenSubscription func() *eventingv1alpha1.Subscription
		isInDeletion      bool
	}{
		{
			name: "Deletion timestamp uninitialized",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = nil
				return sub
			},
			isInDeletion: false,
		},
		{
			name: "Deletion timestamp is zero",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				zero := metav1.Time{}
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = &zero
				return sub
			},
			isInDeletion: false,
		},
		{
			name: "Deletion timestamp is set to a useful time",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				newTime := metav1.NewTime(time.Now())
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = &newTime
				return sub
			},
			isInDeletion: true,
		},
	}
	g := NewGomegaWithT(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			givenSubscription := tt.givenSubscription()
			g.Expect(isInDeletion(givenSubscription)).To(Equal(tt.isInDeletion))
		})
	}
}
func Test_replaceStatusCondition(t *testing.T) {
	var tests = []struct {
		name              string
		giveSubscription  *eventingv1alpha1.Subscription
		giveCondition     eventingv1alpha1.Condition
		wantStatusChanged bool
		wantStatus        *eventingv1alpha1.SubscriptionStatus
		wantReady         bool
	}{
		{
			name: "Updating a condition marks the status as changed",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				subscription.Status.InitializeConditions()
				return subscription
			}(),
			giveCondition: func() eventingv1alpha1.Condition {
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, "")
			}(),
			wantStatusChanged: true,
			wantReady:         false,
		},
		{
			name: "All conditions true means status is ready",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark all conditions as true
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionTrue,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionTrue,
					},
				}
				return subscription
			}(),
			giveCondition: func() eventingv1alpha1.Condition {
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, "")
			}(),
			wantStatusChanged: true, // readiness changed
			wantReady:         true, // all conditions are true
		},
	}

	g := NewGomegaWithT(t)
	r := Reconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.giveSubscription
			condition := tt.giveCondition
			statusChanged := r.replaceStatusCondition(subscription, condition)
			g.Expect(statusChanged).To(BeEquivalentTo(tt.wantStatusChanged))
			g.Expect(subscription.Status.Conditions).To(ContainElement(condition))
			g.Expect(subscription.Status.Ready).To(BeEquivalentTo(tt.wantReady))
		})
	}
}

func Test_syncConditionSubscribed(t *testing.T) {
	currentTime := metav1.Now()

	var tests = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenIsSubscribed bool
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionSubscribed conditions as true
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
				}
				return subscription
			}(),
			givenIsSubscribed: false,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionSubscribed,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionFalse,
				Reason:             eventingv1alpha1.ConditionReasonSubscriptionCreationFailed,
			},
		},
		{
			name: "should replace condition with status true",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionSubscribed conditions as false
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
				}
				return subscription
			}(),
			givenIsSubscribed: true,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionSubscribed,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionTrue,
				Reason:             eventingv1alpha1.ConditionReasonSubscriptionCreated,
				Message:            "BEB-subscription-name=some-namef73aa86661706ae6ba5acf1d32821ce318051d0e",
			},
		},
	}

	g := NewGomegaWithT(t)
	r := Reconciler{
		nameMapper: handlers.NewBEBSubscriptionNameMapper(domain, handlers.MaxBEBSubscriptionNameLength),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.givenSubscription

			// sync condition
			r.syncConditionSubscribed(subscription, tt.givenIsSubscribed)

			// check if the condition was replaced correctly
			newCondition := subscription.Status.FindCondition(tt.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())

			// equate the timestamp, so we only compare the other fields in condition
			newCondition.LastTransitionTime = tt.wantCondition.LastTransitionTime
			g.Expect(newCondition).To(Equal(&tt.wantCondition))
		})
	}
}

func Test_syncConditionSubscriptionActive(t *testing.T) {
	currentTime := metav1.Now()

	var tests = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenIsSubscribed bool
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionSubscriptionActive conditions as true
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
				}
				return subscription
			}(),
			givenIsSubscribed: false,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionSubscriptionActive,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionFalse,
				Reason:             eventingv1alpha1.ConditionReasonSubscriptionNotActive,
			},
		},
		{
			name: "should replace condition with status true",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionSubscriptionActive conditions as false
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
				}
				return subscription
			}(),
			givenIsSubscribed: true,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionSubscriptionActive,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionTrue,
				Reason:             eventingv1alpha1.ConditionReasonSubscriptionActive,
			},
		},
	}

	g := NewGomegaWithT(t)

	logger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf(`failed to initiate logger, %v`, err)
	}

	r := Reconciler{
		nameMapper: handlers.NewBEBSubscriptionNameMapper(domain, handlers.MaxBEBSubscriptionNameLength),
		logger:     logger,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.givenSubscription
			log := utils.LoggerWithSubscription(r.namedLogger(), subscription)

			// sync condition
			r.syncConditionSubscriptionActive(subscription, tt.givenIsSubscribed, log)

			// check if the condition was replaced correctly
			newCondition := subscription.Status.FindCondition(tt.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())

			// equate the timestamp, so we only compare the other fields in condition
			newCondition.LastTransitionTime = tt.wantCondition.LastTransitionTime
			g.Expect(newCondition).To(Equal(&tt.wantCondition))
		})
	}
}

func Test_syncConditionWebhookCallStatus(t *testing.T) {
	currentTime := metav1.Now()

	var tests = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenIsSubscribed bool
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false if it throws error to check lastDelivery",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionWebhookCallStatus conditions as true
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
					{
						Type:               eventingv1alpha1.ConditionWebhookCallStatus,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionTrue,
					},
				}
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   "invalid",
					LastFailedDelivery:       "invalid",
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			givenIsSubscribed: false,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionWebhookCallStatus,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionFalse,
				Reason:             eventingv1alpha1.ConditionReasonWebhookCallStatus,
				Message:            `parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`,
			},
		},
		{
			name: "should replace condition with status false if lastDelivery was not okay",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionWebhookCallStatus conditions as false
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
					{
						Type:               eventingv1alpha1.ConditionWebhookCallStatus,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
				}
				// set EmsSubscriptionStatus
				// LastFailedDelivery is latest
				subscription.Status.EmsSubscriptionStatus = eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   time.Now().Format(time.RFC3339),
					LastFailedDelivery:       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					LastFailedDeliveryReason: "abc",
				}
				return subscription
			}(),
			givenIsSubscribed: true,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionWebhookCallStatus,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionFalse,
				Reason:             eventingv1alpha1.ConditionReasonWebhookCallStatus,
				Message:            "abc",
			},
		},
		{
			name: "should replace condition with status true if lastDelivery was okay",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark ConditionWebhookCallStatus conditions as false
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
					{
						Type:               eventingv1alpha1.ConditionWebhookCallStatus,
						LastTransitionTime: currentTime,
						Status:             corev1.ConditionFalse,
					},
				}
				// set EmsSubscriptionStatus
				// LastSuccessfulDelivery is latest
				subscription.Status.EmsSubscriptionStatus = eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					LastFailedDelivery:       time.Now().Format(time.RFC3339),
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			givenIsSubscribed: true,
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionWebhookCallStatus,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionTrue,
				Reason:             eventingv1alpha1.ConditionReasonWebhookCallStatus,
				Message:            "",
			},
		},
	}

	g := NewGomegaWithT(t)
	logger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf(`failed to initiate logger, %v`, err)
	}

	r := Reconciler{
		logger: logger,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.givenSubscription

			// sync condition
			r.syncConditionWebhookCallStatus(subscription)

			// check if the condition was replaced correctly
			newCondition := subscription.Status.FindCondition(tt.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())

			// equate the timestamp, so we only compare the other fields in condition
			newCondition.LastTransitionTime = tt.wantCondition.LastTransitionTime
			g.Expect(newCondition).To(Equal(&tt.wantCondition))
		})
	}
}
