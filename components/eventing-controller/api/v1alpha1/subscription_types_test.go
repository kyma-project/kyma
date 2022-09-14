//go:build unit
// +build unit

package v1alpha1_test

import (
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/require"

	api "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	orderProcessedEventType = "prefix.testapp1023.order.processed.v1"
	orderCreatedEventType   = "prefix.testapp1023.order.created.v1"
)

func TestBEBFilters_Deduplicate(t *testing.T) {
	filter1 := &api.BEBFilter{
		EventSource: &api.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &api.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	filter2 := &api.BEBFilter{
		EventSource: &api.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &api.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderProcessedEventType,
		},
	}
	filter3 := &api.BEBFilter{
		EventSource: &api.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "/external/system/id",
		},
		EventType: &api.Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	tests := []struct {
		caseName  string
		input     *api.BEBFilters
		expected  *api.BEBFilters
		expectErr bool
	}{
		{
			caseName:  "Only one filter",
			input:     &api.BEBFilters{Dialect: "beb", Filters: []*api.BEBFilter{filter1}},
			expected:  &api.BEBFilters{Dialect: "beb", Filters: []*api.BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters with duplicate",
			input:     &api.BEBFilters{Dialect: "nats", Filters: []*api.BEBFilter{filter1, filter1}},
			expected:  &api.BEBFilters{Dialect: "nats", Filters: []*api.BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters without duplicate",
			input:     &api.BEBFilters{Filters: []*api.BEBFilter{filter1, filter2, filter3}},
			expected:  &api.BEBFilters{Filters: []*api.BEBFilter{filter1, filter2, filter3}},
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
		inputConf      *api.SubscriptionConfig
		inputDefaults  *env.DefaultSubscriptionConfig
		expectedOutput *api.SubscriptionConfig
	}{
		{
			caseName:       "nil input config",
			inputConf:      nil,
			inputDefaults:  defaultConf,
			expectedOutput: &api.SubscriptionConfig{MaxInFlightMessages: 4},
		},
		{
			caseName:       "default is overridden",
			inputConf:      &api.SubscriptionConfig{MaxInFlightMessages: 10},
			inputDefaults:  defaultConf,
			expectedOutput: &api.SubscriptionConfig{MaxInFlightMessages: 10},
		},
		{
			caseName:       "provided input is invalid",
			inputConf:      &api.SubscriptionConfig{MaxInFlightMessages: 0},
			inputDefaults:  defaultConf,
			expectedOutput: &api.SubscriptionConfig{MaxInFlightMessages: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			got := api.MergeSubsConfigs(tt.inputConf, tt.inputDefaults)
			if !reflect.DeepEqual(got, tt.expectedOutput) {
				t.Errorf("MergeSubsConfigs() got = %v, want = %v", got, tt.expectedOutput)
			}
		})
	}
}

func TestInitializeCleanEventTypes(t *testing.T) {
	s := api.Subscription{}
	require.Nil(t, s.Status.CleanEventTypes)

	s.Status.InitializeCleanEventTypes()
	require.NotNil(t, s.Status.CleanEventTypes)
}
