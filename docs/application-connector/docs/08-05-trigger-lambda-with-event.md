---
title: Trigger a lambda with events
type: Tutorials
---

This guide shows how to create a simple lambda function and trigger it with an event.


## Prerequisites

- An Application (App) bound to the `production` Namespace
- Client certificates generated for the connected App.


## Steps

1. Register a service with the following specification to the desired App.

>**NOTE:** See [this](#getting-started-get-the-client-certificate) Getting Started Guide to learn how to register a service.
```json
{
  "name": "my-service",
  "provider": "myCompany",
  "Identifier": "identifier",
  "description": "This is some service",
  "events": {
    "spec": {
      "asyncapi": "1.0.0",
      "info": {
        "title": "Example Events",
        "version": "1.0.0",
        "description": "Description of all the example events"
      },
      "baseTopic": "example.events.com",
      "topics": {
        "exampleEvent.v1": {
          "subscribe": {
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
```

2. Get the `externalName` of the Service Class of the registered service.
```
kubectl -n production get serviceclass {SERVICE_ID}  -o jsonpath='{.spec.externalName}'
```

3. Create a Service Instance for the registered service.
```
cat <<EOF | kubectl apply -f -
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: my-service-instance-name
  namespace: production
spec:
  serviceClassExternalName: {EXTERNAL_NAME}
EOF
```

4. Create a sample lambda function which sends a request to `http://httpbin.org/uuid`. A successful response logs a `Response acquired successfully! Uuid: {RECEIVED_UUID}` message. To create and register the lambda function in the `production` Namespace, run:
```
cat <<EOF | kubectl apply -f -
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: my-lambda
  namespace: production
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
      function: my-lambda
  timeout: ""
  topic: exampleEvent
EOF
```

5. Create a Subscription to allow events to trigger the lambda function.
```
cat <<EOF | kubectl apply -f -
apiVersion: eventing.kyma.cx/v1alpha1
kind: Subscription
metadata:
  labels:
    Function: my-lambda
  name: lambda-my-lambda-exampleevent-v1
  namespace: production
spec:
  endpoint: http://my-lambda.production:8080/
  event_type: exampleEvent
  event_type_version: v1
  include_subscription_name_header: true
  max_inflight: 400
  push_request_timeout_ms: 2000
  source_id: {APP_NAME}
EOF
```

6. Send an event to trigger the created lambda.
  - On a cluster:
    ```
    curl -X POST https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events -k --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -d \
    '{
        "event-type": "exampleEvent",
        "event-type-version": "v1",
        "event-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
        "event-time": "2018-10-16T15:00:00Z",
        "data": "some data"
    }'
    ```
  - On a local deployment:
    ```
    curl -X POST https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/events -k --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -d \
    '{
        "event-type": "exampleEvent",
        "event-type-version": "v1",
        "event-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
        "event-time": "2018-10-16T15:00:00Z",
        "data": "some data"
    }'
    ```

7. Check the logs of the lambda function to see if it was triggered. Every time an event successfully triggers the function, this message appears in the logs: `Response acquired successfully! Uuid: {RECEIVED_UUID}`. Run this command:
```
kubectl -n production logs "$(kubectl -n production get po -l function=my-lambda -o jsonpath='{.items[0].metadata.name}')" -c my-lambda | grep "Response acquired successfully! Uuid: "
```
