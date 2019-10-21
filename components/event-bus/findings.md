# Findings

* new v2 endpoints only accept Content-Type "application/cloudevents+json"
* headers are only parsed if encoding is `binary` currently we always use `structured`
* trace headers: `x-` added by envoy sidecar
* CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z') or digits ('0' to '9') from the ASCII character set, and MUST begin with a lower-case letter. Attribute names SHOULD be descriptive and terse, and SHOULD NOT exceed 20 characters in length.


Test case stuff:
* ```curl -v  -H "ce-eventtypeversion: ichbringihnum" -H "ce-specversion: 0.3" -H "ce-type: com.example.someevent" -H "ce-source: mycontext" -H "ce-id: aaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" -d '{ "data" : "<much wow=\"xml\"/>" }' -H "Content-Type: application/json" -X POST http://localhost:8080/v2/events``` 

* [kyma doc](https://github.com/kyma-project/kyma/blob/master/docs/event-bus/03-03-service-programming-model.md) says to use CE binary encoding
  => Content-Type not explicitely defined
* e2e test (`e2e-tester.go`) uses CE structured with `application/json`
  => wrong Content-Type
* test in event-publish only used CE structured encoding beforehand uses CE structured with `application/json`
  => wrong Content-Type
 
 
 ## ToDo
 - tests for metrics
 - tests for tracing
 - do same changes to event-service as already done to event-bus
 - nachtmaar: do we still need validation of eventtypeversion? Before: `AllowedEventTypeVersionChars = `^[a-zA-Z0-9]+$`

 ## Changes
 * new /v2 endpoint validates event id according to [CE standard](https://github.com/cloudevents/spec/blob/master/json-format.md)
   * not with regexp `AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`` anymore
