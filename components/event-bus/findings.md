# Findings

* new v2 endpoints only accept Content-Type "application/cloudevents+json"
* headers are only parsed if encoding is `binary` currently we always use `structured`
* trace headers: `x-` added by envoy sidecar
* CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z') or digits ('0' to '9') from the ASCII character set, and MUST begin with a lower-case letter. Attribute names SHOULD be descriptive and terse, and SHOULD NOT exceed 20 characters in length.


Test case stuff:
* ```curl -v  -H "ce-eventtypeversion: ichbringihnum" -H "ce-specversion: 0.3" -H "ce-type: com.example.someevent" -H "ce-source: mycontext" -H "ce-id: aaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" -d '{ "data" : "<much wow=\"xml\"/>" }' -H "Content-Type: application/json" -X POST http://localhost:8080/v2/events``` 

