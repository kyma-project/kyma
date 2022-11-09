package v1alpha2_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	subName = "sub"
	subNs   = "test"
)

func Test_Default(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name     string
		givenSub *v1alpha2.Subscription
		wantSub  *v1alpha2.Subscription
	}

	testCases := []TestCase{
		{
			name: "Add TypeMatching Standard and default MaxInFlightMessages value",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
			),
			wantSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
		},
		{
			name: "Add TypeMatching Standard only",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("20"),
			),
			wantSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingStandard(),
			),
		},
		{
			name: "Add default MaxInFlightMessages value only",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingExact(),
			),
			wantSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingExact(),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.givenSub.Default()
			require.Equal(t, testCase.givenSub, testCase.wantSub)
		})
	}
}

func Test_validateSubscription(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name     string
		givenSub *v1alpha2.Subscription
		wantErr  error
	}

	testCases := []TestCase{
		{
			name: "A valid subscription should not have errors",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: nil,
		},
		{
			name: "empty source and TypeMatching Standard should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SourcePath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "valid source and TypeMatching Standard should not return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: nil,
		},
		{
			name: "empty source and TypeMatching Exact should not return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingExact(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: nil,
		},
		{
			name: "nil types field should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "empty types field should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "duplicate types should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{testingv2.OrderCreatedV1Event, testingv2.OrderCreatedV1Event}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.DuplicateTypesErrDetail)}),
		},
		{
			name: "empty event type should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{testingv2.OrderCreatedV1Event, ""}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.LengthErrDetail)}),
		},
		{
			name: "lower than min segments should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{"order"}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.MinSegmentErrDetail)}),
		},
		{
			name: "invalid prefix should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{v1alpha2.InvalidPrefix}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.InvalidPrefixErrDetail)}),
		},
		{
			name: "invalid prefix with exact should not return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{v1alpha2.InvalidPrefix}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: nil,
		},
		{
			name: "invalid maxInFlight value should return error",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("invalid"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.ConfigPath,
					subName, v1alpha2.StringIntErrDetail)}),
		},
		{
			name: "multiple errors should be reported if exists",
			givenSub: testingv2.NewSubscription(subName, subNs,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("invalid"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SourcePath,
					subName, v1alpha2.EmptyErrDetail),
					v1alpha2.MakeInvalidFieldError(v1alpha2.ConfigPath,
						subName, v1alpha2.StringIntErrDetail)}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := testCase.givenSub.ValidateSubscription()
			require.Equal(t, err, testCase.wantErr)
		})
	}
}
