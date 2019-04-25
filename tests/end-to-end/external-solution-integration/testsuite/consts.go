package testsuite

const (
	appName     = "e2e-test-app"
	accessLabel = "e2e-test-app-label"

	integrationNamespace = "kyma-integration"
	productionNamespace  = "production"

	lambdaEndpoint = "http://e2e-lambda.production:8080"
	eventType      = "exampleevent"
	eventVersion   = "v1"

	serviceInstanceName = "e2e-test-app-si"
	serviceInstanceID   = "e2e-test-app-si-id"

	serviceProvider         = "e2e"
	serviceName             = "e2e-test-app-svc"
	serviceDescription      = "e2e testing app"
	serviceShortDescription = "e2e testing app"
	serviceIdentifier       = "e2e-test-app-svc-id"
	serviceEventsSpec       = `{
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
