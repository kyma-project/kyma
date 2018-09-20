package common

import (
	"testing"
)

func TestEventDetails_Encode(t *testing.T) {
	type fields struct {
		eventType        string
		eventTypeVersion string
		sourceID         string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "event1",
			fields: fields{
				sourceID:         "prod.local.kyma.commerce.ec",
				eventType:        "order.created",
				eventTypeVersion: "v1",
			},
			want: `prod\.local\.kyma\.commerce\.ec.order\.created.v1`},
		{name: "event2",
			fields: fields{
				sourceID:         "prod.com.local.kyma.ec",
				eventType:        "order.created",
				eventTypeVersion: "v1",
			},
			want: `prod\.com\.local\.kyma\.ec.order\.created.v1`},
		{name: "event3",
			fields: fields{
				sourceID:         "local.kyma.commerce.prod.local.kyma.commerce.ec",
				eventType:        "order.created.v1",
				eventTypeVersion: "",
			},
			want: `local\.kyma\.commerce\.prod\.local\.kyma\.commerce\.ec.order\.created\.v1.`},
		{name: "event4",
			fields: fields{
				sourceID:         `prod\.com.ec`,
				eventType:        `order\.created`,
				eventTypeVersion: "v1",
			},
			want: `prod\\\.com\.ec.order\\\.created.v1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EventDetails{
				eventType:        tt.fields.eventType,
				eventTypeVersion: tt.fields.eventTypeVersion,
				sourceID:         tt.fields.sourceID,
			}
			if got := e.Encode(); got != tt.want {
				t.Errorf("\nat %v EventDetails.Encode() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
