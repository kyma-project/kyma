package testing

const (
	StructuredCloudEventPayloadWithoutID = `{
		   "type":"someType",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	StructuredCloudEventPayloadWithoutType = `{
		   "id":"00000",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	StructuredCloudEventPayloadWithoutSpecVersion = `{
		   "id":"00000",
		   "type":"someType",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	StructuredCloudEventPayloadWithoutSource = `{
		   "id":"00000",
		   "type":"someType",
		   "specversion":"1.0",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	StructuredCloudEventPayload = `{
		   "id":"00000",
		   "type":"someType",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	BinaryCloudEventPayload = `"{\"foo\":\"bar\"}"`

	ValidLegacyEventPayloadWithEventId = `{
			"event-id": "8945ec08-256b-11eb-9928-acde48001122",
		    "event-type":"someType",
		    "event-type-version":"v0",
			"event-time": "2020-04-02T21:37:00Z",
		    "data": {
				"foo": "bar"
			}
		}`

	ValidLegacyEventPayloadWithoutEventId = `{
		    "event-type":"someType",
		    "event-type-version":"v0",
			"event-time": "2020-04-02T21:37:00Z",
		    "data": {
				"foo": "bar"
			}
		}`
	LegacyEventPayloadWithInvalidEventId = `{
			"event-id":"foo-bar",
		    "event-type":"someType",
		    "event-type-version":"v0",
			"event-time": "2020-04-02T21:37:00Z",
		    "data": {
				"foo": "bar"
			}
		}`

	LegacyEventPayloadWithoutEventTime = `{
		    "event-id":"8945ec08-256b-11eb-9928-acde48001122",
			"event-type":"someType",
		    "event-type-version":"v0",
		    "data": {
				"foo": "bar"
			}
		}`

	LegacyEventPayloadWithoutEventType = `{
		    "event-id":"8945ec08-256b-11eb-9928-acde48001122",
		    "event-type-version":"v0",
			"event-time": "2020-04-02T21:37:00Z",
		    "data": {
				"foo": "bar"
			}
		}`

	LegacyEventPayloadWithInvalidEventTime = `{
		    "event-id":"8945ec08-256b-11eb-9928-acde48001122",
			"event-type":"someType",
		    "event-type-version":"v0",
			"event-time": "some time 10:23",
		    "data": {
				"foo": "bar"
			}
		}`

	LegacyEventPayloadWithWithoutEventVersion = `{
		    "event-id":"8945ec08-256b-11eb-9928-acde48001122",
			"event-type":"someType",
			"event-time": "2020-04-02T21:37:00Z",
		    "data": {
				"foo": "bar"
			}
		}`

	ValidLegacyEventPayloadWithoutData = `{
			"event-id": "8945ec08-256b-11eb-9928-acde48001122",
		    "event-type":"someType",
		    "event-type-version":"v0",
			"event-time": "2020-04-02T21:37:00Z"
		}`
)
