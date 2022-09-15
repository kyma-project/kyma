//go:build unit
// +build unit

package v1alpha1_test

import (
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/require"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	orderProcessedEventType = "prefix.testapp1023.order.processed.v1"
	orderCreatedEventType   = "prefix.testapp1023.order.created.v1"
)

func TestBEBFilters_Deduplicate(t *testing.T) {
	filter1 := &eventingv1alpha1.BEBFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	filter2 := &eventingv1alpha1.BEBFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderProcessedEventType,
		},
	}
	filter3 := &eventingv1alpha1.BEBFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "/external/system/id",
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	tests := []struct {
		caseName  string
		input     *eventingv1alpha1.BEBFilters
		expected  *eventingv1alpha1.BEBFilters
		expectErr bool
	}{
		{
			caseName:  "Only one filter",
			input:     &eventingv1alpha1.BEBFilters{Dialect: "beb", Filters: []*eventingv1alpha1.BEBFilter{filter1}},
			expected:  &eventingv1alpha1.BEBFilters{Dialect: "beb", Filters: []*eventingv1alpha1.BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters with duplicate",
			input:     &eventingv1alpha1.BEBFilters{Dialect: "nats", Filters: []*eventingv1alpha1.BEBFilter{filter1, filter1}},
			expected:  &eventingv1alpha1.BEBFilters{Dialect: "nats", Filters: []*eventingv1alpha1.BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters without duplicate",
			input:     &eventingv1alpha1.BEBFilters{Filters: []*eventingv1alpha1.BEBFilter{filter1, filter2, filter3}},
			expected:  &eventingv1alpha1.BEBFilters{Filters: []*eventingv1alpha1.BEBFilter{filter1, filter2, filter3}},
			expectErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			got, err := tt.input.Deduplicate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Deduplicate() error = %v, expectErr %v", err, tt.expected)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Deduplicate() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeSubsConfigs(t *testing.T) {
	defaultConf := &env.DefaultSubscriptionConfig{MaxInFlightMessages: 4}
	tests := []struct {
		caseName       string
		inputConf      *eventingv1alpha1.SubscriptionConfig
		inputDefaults  *env.DefaultSubscriptionConfig
		expectedOutput *eventingv1alpha1.SubscriptionConfig
	}{
		{
			caseName:       "nil input config",
			inputConf:      nil,
			inputDefaults:  defaultConf,
			expectedOutput: &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 4},
		},
		{
			caseName:       "default is overridden",
			inputConf:      &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 10},
			inputDefaults:  defaultConf,
			expectedOutput: &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 10},
		},
		{
			caseName:       "provided input is invalid",
			inputConf:      &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 0},
			inputDefaults:  defaultConf,
			expectedOutput: &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			got := eventingv1alpha1.MergeSubsConfigs(tt.inputConf, tt.inputDefaults)
			if !reflect.DeepEqual(got, tt.expectedOutput) {
				t.Errorf("MergeSubsConfigs() got = %v, want = %v", got, tt.expectedOutput)
			}
		})
	}
}

func TestInitializeCleanEventTypes(t *testing.T) {
	s := eventingv1alpha1.Subscription{}
	require.Nil(t, s.Status.CleanEventTypes)

	s.Status.InitializeCleanEventTypes()
	require.NotNil(t, s.Status.CleanEventTypes)
}
