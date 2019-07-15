package consts

// All constants used by this test
const (
	AppName     = "e2e-test-app"
	AccessLabel = "e2e-test-app-label"

	LambdaEndpointPattern = "http://e2e-test-app.%s:8080"
	EventType             = "exampleEvent"
	EventVersion          = "v1"

	ServiceInstanceName = "e2e-test-app-si"
	ServiceInstanceID   = "e2e-test-app-si-id"

	ServiceBindingName   = "e2e-test-app-sb"
	ServiceBindingID     = "e2e-test-app-sb-id"
	ServiceBindingSecret = "e2e-test-app-sb"

	ServiceBindingUsageName = "e2e-test-app-sbu"

	ServiceProvider         = "e2e"
	ServiceName             = "e2e-test-app-svc"
	ServiceDescription      = "e2e testing app"
	ServiceShortDescription = "e2e testing app"
	ServiceIdentifier       = "e2e-test-app-svc-id"
	ServiceEventsSpec       = `{
   "asyncapi":"1.0.0",
   "info":{
      "title":"Example Events",
      "version":"1.0.0",
      "description":"Description of all the example events"
   },
   "baseTopic":"example.events.com",
   "topics":{
      "exampleEvent.v1":{
         "subscribe":{
            "summary":"Example event",
            "payload":{
               "type":"object",
               "properties":{
                  "myObject":{
                     "type":"object",
                     "required":[
                        "id"
                     ],
                     "example":{
                        "id":"4caad296-e0c5-491e-98ac-0ed118f9474e"
                     },
                     "properties":{
                        "id":{
                           "title":"Id",
                           "description":"Resource identifier",
                           "type":"string"
                        }
                     }
                  }
               }
            }
         }
      }
   }
}`
)
