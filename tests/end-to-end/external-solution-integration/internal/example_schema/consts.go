package example_schema

// All constants used by this test
const (
	EventType    = "exampleEvent"
	EventVersion = "v1"

	EventsSpec = `{
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
