package testing

const (
	CloudEventWithoutID = `{
		   "type":"someType",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	CloudEventWithoutType = `{
		   "id":"00000",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	CloudEventWithoutSpecVersion = `{
		   "id":"00000",
		   "type":"someType",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	CloudEventWithoutSource = `{
		   "id":"00000",
		   "type":"someType",
		   "specversion":"1.0",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`

	CloudEvent = `{
		   "id":"00000",
		   "type":"someType",
		   "specversion":"1.0",
		   "source":"someSource",
		   "data":"{\"foo\":\"bar\"}",
		   "datacontenttype":"application/json"
		}`
)
