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
	subName      = "sub"
	subNamespace = "test"
	sink         = "https://eventing-nats.test.svc.cluster.local:8080"
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
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
			),
			wantSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
		},
		{
			name: "Add TypeMatching Standard only",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("20"),
			),
			wantSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("20"),
				testingv2.WithTypeMatchingStandard(),
			),
		},
		{
			name: "Add default MaxInFlightMessages value only",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingExact(),
			),
			wantSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithTypeMatchingExact(),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.givenSub.Default()
			require.Equal(t, tc.givenSub, tc.wantSub)
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
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: nil,
		},
		{
			name: "empty source and TypeMatching Standard should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SourcePath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "valid source and TypeMatching Standard should not return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: nil,
		},
		{
			name: "empty source and TypeMatching Exact should not return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingExact(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: nil,
		},
		{
			name: "nil types field should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "empty types field should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "duplicate types should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{testingv2.OrderCreatedV1Event, testingv2.OrderCreatedV1Event}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.DuplicateTypesErrDetail)}),
		},
		{
			name: "empty event type should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{testingv2.OrderCreatedV1Event, ""}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.LengthErrDetail)}),
		},
		{
			name: "lower than min segments should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{"order"}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.MinSegmentErrDetail)}),
		},
		{
			name: "invalid prefix should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{v1alpha2.InvalidPrefix}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.TypesPath,
					subName, v1alpha2.InvalidPrefixErrDetail)}),
		},
		{
			name: "invalid prefix with exact should not return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithTypes([]string{v1alpha2.InvalidPrefix}),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(sink),
			),
			wantErr: nil,
		},
		{
			name: "invalid maxInFlight value should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("invalid"),
				testingv2.WithSink(sink),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.ConfigPath,
					subName, v1alpha2.StringIntErrDetail)}),
		},
		{
			name: "missing sink should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, v1alpha2.EmptyErrDetail)}),
		},
		{
			name: "sink with invalid scheme should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink(subNamespace),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, v1alpha2.MissingSchemeErrDetail)}),
		},
		{
			name: "sink with invalid URL should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink("http://invalid Sink"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, "failed to parse subscription sink URL: "+
						"parse \"http://invalid Sink\": invalid character \" \" in host name")}),
		},
		{
			name: "sink with invalid suffix should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink("https://svc2.test.local"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, v1alpha2.SuffixMissingErrDetail)}),
		},
		{
			name: "sink with invalid suffix and port should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink("https://svc2.test.local:8080"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, v1alpha2.SuffixMissingErrDetail)}),
		},
		{
			name: "sink with invalid number of subdomains should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink("https://svc.cluster.local:8080"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.SinkPath,
					subName, v1alpha2.SubDomainsErrDetail+"svc.cluster.local")}),
		},
		{
			name: "sink with different namespace should return error",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithSource(testingv2.EventSourceClean),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages(v1alpha2.DefaultMaxInFlightMessages),
				testingv2.WithSink("https://eventing-nats.kyma-system.svc.cluster.local"),
			),
			wantErr: apierrors.NewInvalid(
				v1alpha2.GroupKind, subName,
				field.ErrorList{v1alpha2.MakeInvalidFieldError(v1alpha2.NSPath,
					subName, v1alpha2.NSMismatchErrDetail+"kyma-system")}),
		},
		{
			name: "multiple errors should be reported if exists",
			givenSub: testingv2.NewSubscription(subName, subNamespace,
				testingv2.WithTypeMatchingStandard(),
				testingv2.WithEventType(testingv2.OrderCreatedV1Event),
				testingv2.WithMaxInFlightMessages("invalid"),
				testingv2.WithSink(sink),
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
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.givenSub.ValidateSubscription()
			require.Equal(t, err, tc.wantErr)
		})
	}
}
