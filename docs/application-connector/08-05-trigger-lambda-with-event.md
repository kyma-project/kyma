---
title: Trigger a lambda with events
type: Tutorials
---

To create a simple lambda function and trigger it with an event, you must first register a service using the Application Registry that is a part of the Application Connector. This service then sends the event that triggers the lambda. You must create a Service Instance which enables this event in the Namespace. Follow this guide to learn how to do it. 

## Prerequisites

- An Application bound to a Namespace
- Client certificates generated for the connected Application

>**NOTE:** See the respective tutorials to learn how to [create](#tutorials-create-a-new-application) an Application, [get](#tutorials-get-the-client-certificate) the client certificate, and [bind](#tutorials-bind-an-application-to-a-namespace) an Application to a Namespace.

## Steps

1. Export the name of the Namespace to which you bound your Application, and the name of your Application.

   ```bash
   export NAMESPACE={YOUR_NAMESPACE}
   export APP_NAME={YOUR_APPLICATION_NAME}
   ```

2. Register a service with events in the desired Application. Use the example AsyncAPI specification.

   >**NOTE:** See [this](#tutorials-get-the-client-certificate) tutorial to learn how to register a service.

   ```json
   {
     "name": "my-events-service",
     "provider": "myCompany",
     "Identifier": "identifier",
     "description": "This is some service",
     "events": {
       "spec": {
         "asyncapi": "2.0.0",
         "info": {
           "title": "Example Events",
           "version": "2.0.0",
           "description": "Description of all the example events"
         },
         "channels": {
           "example/events/com/exampleEvent/v1": {
             "subscribe": {
               "message": {
                 "summary": "Example event",
                 "payload": {
                   "type": "object",
                   "properties": {
                     "myObject": {
                       "type": "object",
                       "required": [
                         "id"
                       ],
                       "example": {
                         "id": "4caad296-e0c5-491e-98ac-0ed118f9474e"
                       },
                       "properties": {
                         "id": {
                           "title": "Id",
                           "description": "Resource identifier",
                           "type": "string"
                         }
                       }
                     }
                   }
                 }
               }
             }
           }
         }
       }
     }
   }
   ```

3. Expose the `externalName` of the Service Class of the registered service.

   ```bash
   export EXTERNAL_NAME=$(kubectl -n $NAMESPACE get serviceclass {SERVICE_ID}  -o jsonpath='{.spec.externalName}')
   ```

4. Create a Service Instance for the registered service.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: servicecatalog.k8s.io/v1beta1
   kind: ServiceInstance
   metadata:
     name: my-events-service-instance-name
     namespace: $NAMESPACE
   spec:
     serviceClassExternalName: $EXTERNAL_NAME
   EOF
   ```

5. Create and register a lambda function in your Namespace.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: kubeless.io/v1beta1
   kind: Function
   metadata:
     name: my-events-lambda
     namespace: $NAMESPACE
     labels:
       app: my-events-lambda
   spec:
     deployment:
       spec:
         template:
           spec:
             containers:
             - name: ""
               resources: {}
     deps: |-
       {
           "name": "example-1",
           "version": "0.0.1",
           "dependencies": {
             "request": "^2.85.0"
           }
       }
     function: |-
       const request = require('request');

       module.exports = { main: function (event, context) {
           return new Promise((resolve, reject) => {
               const url = \`http://httpbin.org/uuid\`;
               const options = {
                   url: url,
               };

               sendReq(url, resolve, reject)
           })
       } }

       function sendReq(url, resolve, reject) {
           request.get(url, { json: true }, (error, response, body) => {
               if(error){
                   resolve(error);
               }
                 console.log("Response acquired successfully! Uuid: " + response.body.uuid);
               resolve(response);
           })
       }
     function-content-type: text
     handler: handler.main
     horizontalPodAutoscaler:
       spec:
         maxReplicas: 0
     runtime: nodejs8
     service:
       ports:
       - name: http-function-port
         port: 8080
         protocol: TCP
         targetPort: 8080
       selector:
         created-by: kubeless
         function: my-events-lambda
     timeout: ""
     topic: exampleEvent
   EOF
   ```

6. Create a Subscription to allow events to trigger the lambda function.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: eventing.kyma-project.io/v1alpha1
   kind: Subscription
   metadata:
     labels:
       Function: my-events-lambda
     name: lambda-my-events-lambda-exampleevent-v1
     namespace: $NAMESPACE
   spec:
     endpoint: http://my-events-lambda.$NAMESPACE:8080/
     event_type: exampleevent
     event_type_version: v1
     include_subscription_name_header: true
     source_id: $APP_NAME
   EOF
   ```

7. Send an event to trigger the created lambda.

   ```bash
   curl -X POST https://gateway.{CLUSTER_DOMAIN}/$APP_NAME/v1/events -k --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -d \
   '{
       "event-type": "exampleevent",
       "event-type-version": "v1",
       "event-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
       "event-time": "2018-10-16T15:00:00Z",
       "data": "some data"
   }'
   ```

8. Check the logs of the lambda function to see if it was triggered. Every time an event successfully triggers the function, this message appears in the logs: `Response acquired successfully! Uuid: {RECEIVED_UUID}`.

   ```bash
   kubectl -n $NAMESPACE logs "$(kubectl -n $NAMESPACE get po -l function=my-events-lambda -o jsonpath='{.items[0].metadata.name}')" -c my-events-lambda | grep -E "Response acquired successfully! Uuid: [a-f0-9-]+"
   ```
