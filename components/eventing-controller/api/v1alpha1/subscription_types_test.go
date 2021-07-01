package v1alpha1

import (
	"reflect"
	"testing"
)

const (
	orderProcessedEventType = "sap.kyma.testapp1023.order.processed.v1"
	orderCreatedEventType   = "sap.kyma.testapp1023.order.created.v1"
)

func TestBebFilters_Deduplicate(t *testing.T) {
	filter1 := &BebFilter{
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
	filter2 := &BebFilter{
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
	filter3 := &BebFilter{
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
		input     *BebFilters
		expected  *BebFilters
		expectErr bool
	}{
		{
			caseName:  "Only one filter",
			input:     &BebFilters{Dialect: "beb", Filters: []*BebFilter{filter1}},
			expected:  &BebFilters{Dialect: "beb", Filters: []*BebFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters with duplicate",
			input:     &BebFilters{Dialect: "nats", Filters: []*BebFilter{filter1, filter1}},
			expected:  &BebFilters{Dialect: "nats", Filters: []*BebFilter{filter1}},
			expectErr: false,
		},
		{
			caseName:  "Filters without duplicate",
			input:     &BebFilters{Filters: []*BebFilter{filter1, filter2, filter3}},
			expected:  &BebFilters{Filters: []*BebFilter{filter1, filter2, filter3}},
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
