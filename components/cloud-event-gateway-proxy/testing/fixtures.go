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
)
