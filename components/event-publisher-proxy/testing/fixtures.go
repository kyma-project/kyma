package testing

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	cev2 "github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
)

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "testapp_1-0+2=3"

	MessagingNamespace       = "/messaging.namespace"
	MessagingEventTypePrefix = "prefix"

	EventID      = "8945ec08-256b-11eb-9928-acde48001122"
	EventData    = `{\"key\":\"value\"}`
	EventName    = "order.created"
	EventVersion = "v1"

	CloudEventNameAndVersion = EventName + "." + EventVersion
	CloudEventType           = MessagingEventTypePrefix + "." + ApplicationName + "." + CloudEventNameAndVersion
	CloudEventTypeNotClean   = MessagingEventTypePrefix + "." + ApplicationNameNotClean + "." + CloudEventNameAndVersion
	CloudEventSource         = "/default/sap.kyma/id"
	CloudEventSpecVersion    = "1.0"

	LegacyEventTime = "2020-04-02T21:37:00Z"
)

type Event struct {
	id        string
	data      string
	eventType string
}

type CloudEvent struct {
	Event
	specVersion     string
	eventSource     string
	dataContentType string
}

type CloudEventBuilder struct {
	CloudEvent
}

type CloudEventBuilderOpt func(*CloudEventBuilder)

func NewCloudEventBuilder(opts ...CloudEventBuilderOpt) *CloudEventBuilder {
	builder := &CloudEventBuilder{
		CloudEvent{
			Event: Event{
				id:        EventID,
				data:      EventData,
				eventType: CloudEventType,
			},
			specVersion:     CloudEventSpecVersion,
			eventSource:     CloudEventSource,
			dataContentType: internal.ContentTypeApplicationJSON,
		},
	}
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}

func WithCloudEventID(id string) CloudEventBuilderOpt {
	return func(b *CloudEventBuilder) {
		b.id = id
	}
}

func WithCloudEventSource(eventSource string) CloudEventBuilderOpt {
	return func(b *CloudEventBuilder) {
		b.eventSource = eventSource
	}
}

func WithCloudEventSpecVersion(specVersion string) CloudEventBuilderOpt {
	return func(b *CloudEventBuilder) {
		b.specVersion = specVersion
	}
}

func WithCloudEventType(eventType string) CloudEventBuilderOpt {
	return func(b *CloudEventBuilder) {
		b.eventType = eventType
	}
}

func addAsHeaderIfPresent(header http.Header, key, value string) {
	if len(strings.TrimSpace(key)) == 0 || len(strings.TrimSpace(value)) == 0 {
		return
	}
	header.Add(key, value)
}

func (b *CloudEventBuilder) BuildBinary() (string, http.Header) {
	payload := fmt.Sprintf(`"%s"`, b.data)
	headers := make(http.Header)
	addAsHeaderIfPresent(headers, CeIDHeader, b.id)
	addAsHeaderIfPresent(headers, CeTypeHeader, b.eventType)
	addAsHeaderIfPresent(headers, CeSourceHeader, b.eventSource)
	addAsHeaderIfPresent(headers, CeSpecVersionHeader, b.specVersion)
	return payload, headers
}

func (b *CloudEventBuilder) BuildStructured() (string, http.Header) {
	payload := `{
           "id":"` + b.id + `",
           "type":"` + b.eventType + `",
           "source":"` + b.eventSource + `",
           "specversion":"` + b.specVersion + `",
           "datacontenttype":"` + b.dataContentType + `"
        }`
	headers := http.Header{internal.HeaderContentType: []string{internal.ContentTypeApplicationCloudEventsJSON}}
	return payload, headers
}

func (b *CloudEventBuilder) Build(t *testing.T) *cev2.Event {
	e := cev2.New(b.specVersion)
	assert.NoError(t, e.Context.SetID(b.id))
	assert.NoError(t, e.Context.SetType(b.eventType))
	assert.NoError(t, e.Context.SetSource(b.eventSource))
	assert.NoError(t, e.SetData(b.dataContentType, b.data))
	return &e
}

type LegacyEvent struct {
	Event
	eventTime        string
	eventTypeVersion string
}

type LegacyEventBuilder struct {
	LegacyEvent
}

type LegacyEventBuilderOpt func(*LegacyEventBuilder)

func NewLegacyEventBuilder(opts ...LegacyEventBuilderOpt) *LegacyEventBuilder {
	builder := &LegacyEventBuilder{
		LegacyEvent{
			Event: Event{
				id:        EventID,
				data:      EventData,
				eventType: EventName,
			},
			eventTime:        LegacyEventTime,
			eventTypeVersion: EventVersion,
		},
	}
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}

func WithLegacyEventID(id string) LegacyEventBuilderOpt {
	return func(b *LegacyEventBuilder) {
		b.id = id
	}
}

func WithLegacyEventType(eventType string) LegacyEventBuilderOpt {
	return func(b *LegacyEventBuilder) {
		b.eventType = eventType
	}
}

func WithLegacyEventTime(eventTime string) LegacyEventBuilderOpt {
	return func(b *LegacyEventBuilder) {
		b.eventTime = eventTime
	}
}

func WithLegacyEventTypeVersion(eventTypeVersion string) LegacyEventBuilderOpt {
	return func(b *LegacyEventBuilder) {
		b.eventTypeVersion = eventTypeVersion
	}
}

func WithLegacyEventData(data string) LegacyEventBuilderOpt {
	return func(b *LegacyEventBuilder) {
		b.data = data
	}
}

func (b *LegacyEventBuilder) Build() (string, http.Header) {
	payload := `{
            "data": "` + b.data + `",
            "event-id": "` + b.id + `",
            "event-type":"` + b.eventType + `",
            "event-time": "` + b.eventTime + `",
            "event-type-version":"` + b.eventTypeVersion + `"
        }`
	headers := http.Header{internal.HeaderContentType: []string{internal.ContentTypeApplicationJSON}}
	return payload, headers
}
