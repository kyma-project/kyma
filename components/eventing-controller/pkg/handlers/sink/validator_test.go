package sink

import (
	"context"
	"strings"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSinkValidator(t *testing.T) {
	// given
	namespaceName := "test"
	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	ctx := context.Background()
	recorder := &record.FakeRecorder{}
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	sinkValidator := NewValidator(ctx, fakeClient, recorder, defaultLogger)

	testCases := []struct {
		name                  string
		givenSubscriptionSink string
		givenSvcNameToCreate  string
		wantErrString         string
	}{
		{
			name:                  "With invalid scheme",
			givenSubscriptionSink: "invalid Sink",
			wantErrString:         MissingSchemeErrMsg,
		},
		{
			name:                  "With invalid URL",
			givenSubscriptionSink: "http://invalid Sink",
			wantErrString:         "not able to parse sink url with error",
		},
		{
			name:                  "With invalid suffix",
			givenSubscriptionSink: "https://svc2.test.local",
			wantErrString:         "sink does not contain suffix",
		},
		{
			name:                  "With invalid suffix and port",
			givenSubscriptionSink: "https://svc2.test.local:8080",
			wantErrString:         "sink does not contain suffix",
		},
		{
			name:                  "With invalid number of subdomains",
			givenSubscriptionSink: "https://svc.cluster.local:8080", // right suffix but 3 subdomains
			wantErrString:         "sink should contain 5 sub-domains",
		},
		{
			name:                  "With different namespaces in subscription and sink name",
			givenSubscriptionSink: "https://eventing-nats.kyma-system.svc.cluster.local:8080", // sub is in test ns
			wantErrString:         "namespace of subscription: test and the namespace of subscriber: kyma-system are different",
		},
		{
			name:                  "With no existing svc in the cluster",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			wantErrString:         "sink is not a valid cluster local svc, failed with error",
		},
		{
			name:                  "With no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			givenSvcNameToCreate:  "test", // wrong name
			wantErrString:         "sink is not a valid cluster local svc, failed with error",
		},
		{
			name:                  "With correct format but missing scheme",
			givenSubscriptionSink: "eventing-nats.test.svc.cluster.local:8080",
			wantErrString:         MissingSchemeErrMsg,
		},
		{
			name:                  "With no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			givenSvcNameToCreate:  "eventing-nats",
			wantErrString:         "",
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := controllertesting.NewSubscription(
				"foo", namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{}),
				controllertesting.WithStatus(true),
				controllertesting.WithSink(testCase.givenSubscriptionSink),
			)

			// create the service if required for test
			if testCase.givenSvcNameToCreate != "" {
				svc := &corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name:      testCase.givenSvcNameToCreate,
						Namespace: namespaceName,
					},
				}

				err := fakeClient.Create(ctx, svc)
				require.NoError(t, err)
			}

			// when
			// call the defaultSinkValidator function
			err := sinkValidator.Validate(sub)

			// then
			// given error should match expected error
			if testCase.wantErrString == "" {
				require.NoError(t, err)
			} else {
				substringResult := strings.Contains(err.Error(), testCase.wantErrString)
				require.True(t, substringResult)
			}
		})
	}
}
