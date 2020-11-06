package apirule

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler"
)

func Test_isRelevantAPIRuleName(t *testing.T) {
	testCases := []struct {
		givenName string
		wantValue bool
	}{
		{
			givenName: "my-apirule",
			wantValue: false,
		},
		{
			givenName: "webhook",
			wantValue: false,
		},
		{
			givenName: fmt.Sprintf("%sapirule", reconciler.ApiRuleNamePrefix),
			wantValue: true,
		},
	}

	for _, tc := range testCases {
		if got := isRelevantAPIRuleName(tc.givenName); got != tc.wantValue {
			t.Errorf("Test relevant APIRule name: [%s] failed, want: [%v] but got: [%v]",
				tc.givenName, tc.wantValue, got)
		}
	}
}

func Test_hasRelevantAPIRuleLabels(t *testing.T) {
	testCases := []struct {
		name         string
		givenApiRule *apigatewayv1alpha1.APIRule
		wantValue    bool
	}{
		{
			name:         "APIRule with nil labels",
			givenApiRule: &apigatewayv1alpha1.APIRule{},
			wantValue:    false,
		},
		{
			name:         "APIRule with empty labels",
			givenApiRule: &apigatewayv1alpha1.APIRule{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
			wantValue:    false,
		},
		{
			name: "APIRule with not relevant labels",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"key1": "value1", "key2": "value2"},
				},
			},
			wantValue: false,
		},
		{
			name: "APIRule with relevant labels",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue},
				},
			},
			wantValue: true,
		},
	}

	for _, tc := range testCases {
		if got := hasRelevantAPIRuleLabels(tc.givenApiRule.Labels); got != tc.wantValue {
			t.Errorf("Test: [%s] failed for APIRule with labels: [%v], want: [%v] but got: [%v]",
				tc.name, tc.givenApiRule.Labels, tc.wantValue, got)
		}
	}
}

func Test_computeAPIRuleReadyStatus(t *testing.T) {
	testCases := []struct {
		name            string
		givenApiRule    *apigatewayv1alpha1.APIRule
		wantReadyStatus bool
	}{
		{
			name: "APIRule with nil statuses",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Status: apigatewayv1alpha1.APIRuleStatus{
					APIRuleStatus:        nil,
					VirtualServiceStatus: nil,
					AccessRuleStatus:     nil,
				},
			},
			wantReadyStatus: false,
		},
		{
			name: "APIRule with one status error",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Status: apigatewayv1alpha1.APIRuleStatus{
					APIRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusError,
					},
					VirtualServiceStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
					AccessRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
				},
			},
			wantReadyStatus: false,
		},
		{
			name: "APIRule with one status skipped",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Status: apigatewayv1alpha1.APIRuleStatus{
					APIRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusSkipped,
					},
					VirtualServiceStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
					AccessRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
				},
			},
			wantReadyStatus: false,
		},
		{
			name: "APIRule with all status okay",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Status: apigatewayv1alpha1.APIRuleStatus{
					APIRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
					VirtualServiceStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
					AccessRuleStatus: &apigatewayv1alpha1.APIRuleResourceStatus{
						Code: apigatewayv1alpha1.StatusOK,
					},
				},
			},
			wantReadyStatus: true,
		},
	}

	for _, tc := range testCases {
		if got := computeAPIRuleReadyStatus(tc.givenApiRule); got != tc.wantReadyStatus {
			t.Errorf("Test: [%s] APIRule ready status failed, want: [%v] but got: [%v]",
				tc.name, tc.wantReadyStatus, got)
		}
	}
}

func Test_setSubscriptionStatusExternalSink(t *testing.T) {
	testCases := []struct {
		name              string
		givenApiRule      *apigatewayv1alpha1.APIRule
		givenSubscription *v1alpha1.Subscription
		wantExternalSink  string
		wantError         bool
	}{
		{
			name: "Invalid Subscription sink",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Spec: apigatewayv1alpha1.APIRuleSpec{
					Service: &apigatewayv1alpha1.Service{
						Host: stringRef("some-host.com"),
					},
				},
			},
			givenSubscription: &v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{Name: "test-subscription1", Namespace: "test-namespace"},
				Spec: v1alpha1.SubscriptionSpec{
					Sink: "invalid~sink",
				},
			},
			wantExternalSink: "",
			wantError:        true,
		},
		{
			name: "Invalid APIRule with nil service",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Spec: apigatewayv1alpha1.APIRuleSpec{
					Service: nil,
				},
			},
			givenSubscription: &v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{Name: "test-subscription1", Namespace: "test-namespace"},
				Spec: v1alpha1.SubscriptionSpec{
					Sink: "http://service.namespace/endpoint",
				},
			},
			wantExternalSink: "",
			wantError:        true,
		},
		{
			name: "Invalid APIRule with nil host",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Spec: apigatewayv1alpha1.APIRuleSpec{
					Service: &apigatewayv1alpha1.Service{
						Host: nil,
					},
				},
			},
			givenSubscription: &v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{Name: "test-subscription1", Namespace: "test-namespace"},
				Spec: v1alpha1.SubscriptionSpec{
					Sink: "http://service.namespace/endpoint",
				},
			},
			wantExternalSink: "",
			wantError:        true,
		},
		{
			name: "Valid Subscription sink and valid APIRule service host",
			givenApiRule: &apigatewayv1alpha1.APIRule{
				Spec: apigatewayv1alpha1.APIRuleSpec{
					Service: &apigatewayv1alpha1.Service{
						Host: stringRef("some-host.com"),
					},
				},
			},
			givenSubscription: &v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{Name: "test-subscription2", Namespace: "test-namespace"},
				Spec: v1alpha1.SubscriptionSpec{
					Sink: "http://service.namespace/endpoint",
				},
			},
			wantExternalSink: "https://some-host.com/endpoint",
			wantError:        false,
		},
	}

	for _, tc := range testCases {
		err := setSubscriptionStatusExternalSink(tc.givenSubscription, tc.givenApiRule)
		if tc.wantError && err == nil {
			t.Errorf("Test: [%s] should have failed with error: [%v]", tc.name, err)
			continue
		}
		if !tc.wantError && err != nil {
			t.Errorf("Test: [%s] should have succedded but returned error: [%v]", tc.name, err)
			continue
		}
		if tc.wantExternalSink != tc.givenSubscription.Status.ExternalSink {
			t.Errorf("Test: [%s] with Subscription sink: [%s] and APIRule service host: [%s] failed, want external sink [%s] but got [%s]",
				tc.name, tc.givenSubscription.Spec.Sink, *tc.givenApiRule.Spec.Service.Host, tc.wantExternalSink, tc.givenSubscription.Status.ExternalSink)
		}
	}
}

func stringRef(value string) *string {
	return &value
}
