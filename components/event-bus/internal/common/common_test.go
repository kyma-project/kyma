package common

import (
	"testing"
)

func TestEventDetails_Encode(t *testing.T) {
	type fields struct {
		eventType        string
		eventTypeVersion string
		source           *source
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "event1",
			fields: fields{
				eventType:        "order.created",
				eventTypeVersion: "v1",
				source: &source{
					sourceEnvironment: "prod",
					sourceNamespace:   "local.kyma.commerce",
					sourceType:        "ec",
				},
			},
			want: `prod.com\.sap\.hybris.ec.order\.created.v1`},
		{name: "event2",
			fields: fields{
				eventType:        "order.created",
				eventTypeVersion: "v1",
				source: &source{
					sourceEnvironment: "prod.com",
					sourceNamespace:   "sap.hybris",
					sourceType:        "ec",
				},
			},
			want: `prod\.com.sap\.hybris.ec.order\.created.v1`},
		{name: "event3",
			fields: fields{
				eventType:        "order.created.v1",
				eventTypeVersion: "",
				source: &source{
					sourceEnvironment: "local.kyma.commerce.prod",
					sourceNamespace:   "local.kyma.commerce",
					sourceType:        "ec",
				},
			},
			want: `com\.sap\.hybris\.prod.com\.sap\.hybris.ec.order\.created\.v1.`},
		{name: "event4",
			fields: fields{
				eventType:        `order\.created`,
				eventTypeVersion: "v1",
				source: &source{
					sourceEnvironment: `prod\`,
					sourceNamespace:   `com.sap.\hybris`,
					sourceType:        "ec",
				},
			},
			want: `prod\\.com\.sap\.\\hybris.ec.order\\\.created.v1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EventDetails{
				eventType:        tt.fields.eventType,
				eventTypeVersion: tt.fields.eventTypeVersion,
				source:           tt.fields.source,
			}
			if got := e.Encode(); got != tt.want {
				t.Errorf("\nat %v EventDetails.Encode() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
