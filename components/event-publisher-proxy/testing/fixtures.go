package testing

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "test.app_1-0+2=3"

	// messaging system properties
	MessagingNamespace       = "/messaging.namespace"
	MessagingEventTypePrefix = "sap.kyma"

	// event properties
	EventID                   = "8945ec08-256b-11eb-9928-acde48001122"
	CloudEventType            = MessagingEventTypePrefix + "." + ApplicationName + ".order.created.v1"
	CloudEventTypeNotClean    = MessagingEventTypePrefix + "." + ApplicationNameNotClean + ".order.created.v1"
	CloudEventSource          = "/default/sap.kyma/id"
	CloudEventSpecVersion     = "1.0"
	CloudEventData            = `{\"key\":\"value\"}`
	CloudEventDataContentType = "application/json"
	LegacyEventType           = "order.created"
	LegacyEventTypeVersion    = "v1"
	LegacyEventTime           = "2020-04-02T21:37:00Z"
	LegacyEventData           = `{"key": "value"}`

	StructuredCloudEventPayloadWithoutID = `{
           "type":"` + CloudEventTypeNotClean + `",
           "specversion":"` + CloudEventSpecVersion + `",
           "source":"` + CloudEventSource + `",
           "data":"` + CloudEventData + `",
           "datacontenttype":"` + CloudEventDataContentType + `"
        }`

	StructuredCloudEventPayloadWithoutType = `{
           "id":"` + EventID + `",
           "specversion":"` + CloudEventSpecVersion + `",
           "source":"` + CloudEventSource + `",
           "data":"` + CloudEventData + `",
           "datacontenttype":"` + CloudEventDataContentType + `"
        }`

	StructuredCloudEventPayloadWithoutSpecVersion = `{
           "id":"` + EventID + `",
           "type":"` + CloudEventTypeNotClean + `",
           "source":"` + CloudEventSource + `",
           "data":"` + CloudEventData + `",
           "datacontenttype":"` + CloudEventDataContentType + `"
        }`

	StructuredCloudEventPayloadWithoutSource = `{
           "id":"` + EventID + `",
           "type":"` + CloudEventTypeNotClean + `",
           "specversion":"` + CloudEventSpecVersion + `",
           "data":"` + CloudEventData + `",
           "datacontenttype":"` + CloudEventDataContentType + `"
        }`

	StructuredCloudEventPayload = `{
           "id":"` + EventID + `",
           "type":"` + CloudEventTypeNotClean + `",
           "specversion":"` + CloudEventSpecVersion + `",
           "source":"` + CloudEventSource + `",
           "data":"` + CloudEventData + `",
           "datacontenttype":"` + CloudEventDataContentType + `"
        }`

	ValidLegacyEventPayloadWithEventId = `{
            "event-id": "` + EventID + `",
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "` + LegacyEventTime + `",
            "data": ` + LegacyEventData + `
        }`

	ValidLegacyEventPayloadWithoutEventId = `{
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "` + LegacyEventTime + `",
            "data": ` + LegacyEventData + `
        }`
	LegacyEventPayloadWithInvalidEventId = `{
            "event-id":"foo-bar",
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "` + LegacyEventTime + `",
            "data": ` + LegacyEventData + `
        }`

	LegacyEventPayloadWithoutEventTime = `{
            "event-id": "` + EventID + `",
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "data": ` + LegacyEventData + `
        }`

	LegacyEventPayloadWithoutEventType = `{
            "event-id": "` + EventID + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "` + LegacyEventTime + `",
            "data": ` + LegacyEventData + `
        }`

	LegacyEventPayloadWithInvalidEventTime = `{
            "event-id": "` + EventID + `",
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "some time 10:23",
            "data": ` + LegacyEventData + `
        }`

	LegacyEventPayloadWithWithoutEventVersion = `{
            "event-id": "` + EventID + `",
            "event-type":"` + LegacyEventType + `",
            "event-time": "` + LegacyEventTime + `",
            "data": ` + LegacyEventData + `
        }`

	ValidLegacyEventPayloadWithoutData = `{
            "event-id": "` + EventID + `",
            "event-type":"` + LegacyEventType + `",
            "event-type-version":"` + LegacyEventTypeVersion + `",
            "event-time": "` + LegacyEventTime + `"
        }`
)
