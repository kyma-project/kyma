package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

const (
	orderProcessedEventType = "sap.kyma.testapp1023.order.processed.v1"
	orderCreatedEventType   = "sap.kyma.testapp1023.order.created.v1"
)

func TestBEBFilters_Deduplicate(t *testing.T) {
	filter1 := &BEBFilter{
		EventSource: &Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	filter2 := &BEBFilter{
		EventSource: &Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderProcessedEventType,
		},
	}
	filter3 := &BEBFilter{
		EventSource: &Filter{
			Type:     "exact",
			Property: "source",
			Value:    "/external/system/id",
		},
		EventType: &Filter{
			Type:     "exact",
			Property: "type",
			Value:    orderCreatedEventType,
		},
	}
	tests := []struct {
		caseName  string
		input     *BEBFilters
		expected  *BEBFilters
		expectErr bool
	}{
		{
			caseName:  "Only one filter",
			input:     &BEBFilters{Dialect: "beb", Filters: []*BEBFilter{filter1}},
			expected:  &BEBFilters{Dialect: "beb", Filters: []*BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters with duplicate",
			input:     &BEBFilters{Dialect: "nats", Filters: []*BEBFilter{filter1, filter1}},
			expected:  &BEBFilters{Dialect: "nats", Filters: []*BEBFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters without duplicate",
			input:     &BEBFilters{Filters: []*BEBFilter{filter1, filter2, filter3}},
			expected:  &BEBFilters{Filters: []*BEBFilter{filter1, filter2, filter3}},
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
		inputConf      *SubscriptionConfig
		inputDefaults  *env.DefaultSubscriptionConfig
		expectedOutput *SubscriptionConfig
	}{
		{
			caseName:       "nil input config",
			inputConf:      nil,
			inputDefaults:  defaultConf,
			expectedOutput: &SubscriptionConfig{MaxInFlightMessages: 4},
		},
		{
			caseName:       "default is overridden",
			inputConf:      &SubscriptionConfig{MaxInFlightMessages: 10},
			inputDefaults:  defaultConf,
			expectedOutput: &SubscriptionConfig{MaxInFlightMessages: 10},
		},
		{
			caseName:       "provided input is invalid",
			inputConf:      &SubscriptionConfig{MaxInFlightMessages: 0},
			inputDefaults:  defaultConf,
			expectedOutput: &SubscriptionConfig{MaxInFlightMessages: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			got := MergeSubsConfigs(tt.inputConf, tt.inputDefaults)
			if !reflect.DeepEqual(got, tt.expectedOutput) {
				t.Errorf("MergeSubsConfigs() got = %v, want = %v", got, tt.expectedOutput)
			}
		})
	}
}
