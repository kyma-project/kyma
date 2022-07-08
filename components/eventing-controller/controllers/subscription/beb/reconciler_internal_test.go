package beb

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/pkg/errors"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventinglogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	domain = "domain.com"
)

func Test_isInDeletion(t *testing.T) {
	var testCases = []struct {
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

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			givenSubscription := tt.givenSubscription()
			g.Expect(isInDeletion(givenSubscription)).To(Equal(tt.isInDeletion))
		})
	}
}
func Test_replaceStatusCondition(t *testing.T) {
	var testCases = []struct {
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

	for _, tt := range testCases {
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

func Test_getRequiredConditions(t *testing.T) {
	var emptySubscriptionStatus eventingv1alpha1.SubscriptionStatus
	emptySubscriptionStatus.InitializeConditions()
	expectedConditions := emptySubscriptionStatus.Conditions

	testCases := []struct {
		name                   string
		subscriptionConditions []eventingv1alpha1.Condition
		wantConditions         []eventingv1alpha1.Condition
	}{
		{
			name:                   "should get expected conditions if the subscription has no conditions",
			subscriptionConditions: []eventingv1alpha1.Condition{},
			wantConditions:         expectedConditions,
		},
		{
			name: "should get subscription conditions if the all the expected conditions are present",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
		},
		{
			name: "should get latest conditions Status compared to the expected condition status",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionUnknown},
			},
		},
		{
			name: "should get rid of unwanted conditions in the subscription, if present",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: "Fake Condition Type", Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionUnknown},
			},
		},
	}

	g := NewGomegaWithT(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotConditions := getRequiredConditions(tc.subscriptionConditions, expectedConditions)
			if !eventingv1alpha1.ConditionsEquals(gotConditions, tc.wantConditions) {
				t.Errorf("ShouldUpdateReadyStatus is not valid, want: %v but got: %v", tc.wantConditions, gotConditions)
			}
			g.Expect(len(gotConditions)).To(BeEquivalentTo(len(expectedConditions)))
		})
	}
}

func Test_syncConditionSubscribed(t *testing.T) {
	currentTime := metav1.Now()
	errorMessage := "error message"
	var testCases = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenError        error
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
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
			givenError: errors.New(errorMessage),
			wantCondition: eventingv1alpha1.Condition{
				Type:               eventingv1alpha1.ConditionSubscribed,
				LastTransitionTime: currentTime,
				Status:             corev1.ConditionFalse,
				Reason:             eventingv1alpha1.ConditionReasonSubscriptionCreationFailed,
				Message:            errorMessage,
			},
		},
		{
			name: "should replace condition with status true",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
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
			givenError: nil,
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			subscription := tc.givenSubscription

			// when
			r.syncConditionSubscribed(subscription, tc.givenError)

			// then
			newCondition := subscription.Status.FindCondition(tc.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())
			g.Expect(eventingv1alpha1.ConditionEquals(*newCondition, tc.wantCondition)).To(Equal(true))
		})
	}
}

func Test_syncConditionSubscriptionActive(t *testing.T) {
	currentTime := metav1.Now()

	var testCases = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenIsSubscribed bool
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: "Paused",
				}

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
				Message:            "current subscription status: Paused",
			},
		},
		{
			name: "should replace condition with status true",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{}

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

	logger, err := eventinglogger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf(`failed to initiate logger, %v`, err)
	}

	r := Reconciler{
		nameMapper: handlers.NewBEBSubscriptionNameMapper(domain, handlers.MaxBEBSubscriptionNameLength),
		logger:     logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			subscription := tc.givenSubscription
			log := utils.LoggerWithSubscription(r.namedLogger(), subscription)

			// when
			r.syncConditionSubscriptionActive(subscription, tc.givenIsSubscribed, log)

			// then
			newCondition := subscription.Status.FindCondition(tc.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())
			g.Expect(eventingv1alpha1.ConditionEquals(*newCondition, tc.wantCondition)).To(Equal(true))
		})
	}
}

func Test_syncConditionWebhookCallStatus(t *testing.T) {
	currentTime := metav1.Now()

	var testCases = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		givenIsSubscribed bool
		wantCondition     eventingv1alpha1.Condition
	}{
		{
			name: "should replace condition with status false if it throws error to check lastDelivery",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
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
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
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
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
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
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
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
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
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
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
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
	logger, err := eventinglogger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf(`failed to initiate logger, %v`, err)
	}

	r := Reconciler{
		logger: logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			subscription := tc.givenSubscription

			// when
			r.syncConditionWebhookCallStatus(subscription)

			// then
			newCondition := subscription.Status.FindCondition(tc.wantCondition.Type)
			g.Expect(newCondition).ShouldNot(BeNil())
			g.Expect(eventingv1alpha1.ConditionEquals(*newCondition, tc.wantCondition)).To(Equal(true))
		})
	}
}

func Test_checkStatusActive(t *testing.T) {
	currentTime := time.Now()
	testCases := []struct {
		name         string
		subscription *eventingv1alpha1.Subscription
		wantStatus   bool
		wantError    error
	}{
		{
			name: "should return active since the EmsSubscriptionStatus is active",
			subscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: string(types.SubscriptionStatusActive),
				}
				return subscription
			}(),
			wantStatus: true,
			wantError:  nil,
		},
		{
			name: "should return active if the EmsSubscriptionStatus is active and the FailedActivation time is set",
			subscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.FailedActivation = currentTime.Format(time.RFC3339)
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: string(types.SubscriptionStatusActive),
				}
				return subscription
			}(),
			wantStatus: true,
			wantError:  nil,
		},
		{
			name: "should return not active if the EmsSubscriptionStatus is inactive",
			subscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: string(types.SubscriptionStatusPaused),
				}
				return subscription
			}(),
			wantStatus: false,
			wantError:  nil,
		},
		{
			name: "should return not active if the EmsSubscriptionStatus is inactive and the the FailedActivation time is set",
			subscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.FailedActivation = currentTime.Format(time.RFC3339)
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: string(types.SubscriptionStatusPaused),
				}
				return subscription
			}(),
			wantStatus: false,
			wantError:  nil,
		},
		{
			name: "should error if timed out waiting after retrying",
			subscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				subscription.Status.FailedActivation = currentTime.Add(time.Minute * -1).Format(time.RFC3339)
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					SubscriptionStatus: string(types.SubscriptionStatusPaused),
				}
				return subscription
			}(),
			wantStatus: false,
			wantError:  errors.New("timeout waiting for the subscription to be active: some-name"),
		},
	}

	g := NewGomegaWithT(t)
	r := Reconciler{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, err := r.checkStatusActive(tc.subscription)
			g.Expect(gotStatus).To(BeEquivalentTo(tc.wantStatus))
			if tc.wantError == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(tc.wantError.Error()))
			}
		})
	}
}

func Test_checkLastFailedDelivery(t *testing.T) {
	var testCases = []struct {
		name              string
		givenSubscription *eventingv1alpha1.Subscription
		wantResult        bool
		wantError         error
	}{
		{
			name: "should return false if there is no lastFailedDelivery",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   "",
					LastFailedDelivery:       "",
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			wantResult: false,
			wantError:  nil,
		},
		{
			name: "should return error if LastFailedDelivery is invalid",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   "",
					LastFailedDelivery:       "invalid",
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			wantResult: true,
			wantError:  errors.New(`parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`),
		},
		{
			name: "should return error if LastFailedDelivery is valid but LastSuccessfulDelivery is invalid",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   "invalid",
					LastFailedDelivery:       time.Now().Format(time.RFC3339),
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			wantResult: true,
			wantError:  errors.New(`parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`),
		},
		{
			name: "should return true if last delivery of event was failed",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   time.Now().Format(time.RFC3339),
					LastFailedDelivery:       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			wantResult: true,
			wantError:  nil,
		},
		{
			name: "should return false if last delivery of event was success",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace")
				// set EmsSubscriptionStatus
				subscription.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
					LastSuccessfulDelivery:   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					LastFailedDelivery:       time.Now().Format(time.RFC3339),
					LastFailedDeliveryReason: "",
				}
				return subscription
			}(),
			wantResult: false,
			wantError:  nil,
		},
	}

	g := NewGomegaWithT(t)
	logger, err := eventinglogger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf(`failed to initiate logger, %v`, err)
	}

	r := Reconciler{
		logger: logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := r.checkLastFailedDelivery(tc.givenSubscription)
			g.Expect(result).To(Equal(tc.wantResult))
			if tc.wantError == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(tc.wantError.Error()))
			}
		})
	}
}
