---
title: Set asynchronous communication between Functions
---

This tutorial demonstrates how to connect two Functions asynchronously. It is based on an [In-cluster Eventing example](https://github.com/kyma-project/examples/pull/188).

The example provides a very simple scenario of two Functions, where the first Function accepts the incoming traffic via HTTP, sanitises the payload and publishes the content as an in-cluster event via [Kyma Eventing](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/eventing/).
The second Function is a message receiver. It subscribes to the given event type and stores the payload.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Kyma CLI](https://github.com/kyma-project/cli)
- Kyma installed locally or on a cluster

## Steps

1. Export the `KUBECONFIG` variable:
   ```bash
   export KUBECONGIG={KUBECONFIG_PATH}
   ```
2. Create two folders in your project - `emitter` and `receiver`.
3. Go to your `emitter` folder and run the `init` Kyma CLI command to initialise a scaffold for your first Function:

   ```bash
   kyma init function
    ```

  The `init` command creates these files in your workspace folder:

  - `config.yaml`	with the Function's configuration

>**NOTE:** See the detailed description of all fields available in the [`config.yaml` file](../../05-technical-reference/svls-06-function-configuration-file.md).

  - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
  - `package.json` with the Function's dependencies

4. In the `config.yaml` file configure an API Rule to expose your Function to the incoming traffic over HTTP. Enter the subdomain name as the `host` property:

    ```yaml
    apiRules:
      - name: incoming-http-trigger
        service:
          host: incoming
        rules:
          - methods:
              - GET
            accessStrategies:
              - handler: allow
    ```

5. Provide your Function logic in the `handler.js` file. In the following example you do not find an actual sanitisation logic, `sanitise` Function is just a placeholder:

   ```js
   const { v4: uuidv4 } = require('uuid');
   module.exports - {
       main: function (event, context) {
           let sanitsedData = sanitise(event.data)
           var eventOut=event.buildResponseCloudEvent(uuidv4(), "sap.kyma.custom.acme.payload.sanitised.v1", sanitisedData);
           eventOut.source="kyma"
           eventOut.specversion="1.0"
           event.publishCloudEvent(eventOut);
           console.log(`Payload pushed to sap.kyma.custom.acme.payload.sanitised.v1`,eventOut)
           return eventOut;
       }
   }
   let sanitise = (data)=>{
       console.log(`sanitising data...`)
       console.log(data)
       return data
   }
   ```
   The `sap.kyma.custom.acme.payload.sanitised.v1` is an exemplary event type declared by the emitter Function when publishing events. You can choose a different one that better suits your use case. Keep in mind the constraints described on the [Event names](https://kyma-project.io/docs/kyma/latest/05-technical-reference/evnt-01-event-names) page. The receiver subscribes to the event type to consume the events.

   Please note that event object provides convinience functions to build and publish events. To send the event you need to build the Cloud Event. To learn more please visit the [Function's Specification](https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-08-function-specification#event-object-sdk) page. In addition, your `eventOut.source` key needs to point at `“kyma”` to use Kyma In-cluster Eventing.

6. Apply your emitter Function:

    ```bash
    kyma apply function
    ```
   Having applied it, your Function is built and deployed in Kyma runtime. Kyma will expose it via API Rule. Any incoming payload would be processed by your emitter Function. It sends the sanitised content to whatever workload that subscribes to selected event type - in our case - receiver Function.

7. Test the first Function. Send a paylod and see if your HTTP traffic is accepted:

      ```bash
      export KYMA_DOMAIN=.... # export the variable for you Kyma domain
      
      curl -X POST https://incoming.${KYMA_DOMAIN}
      -H 'Content-Type: application/json'
      -d '{"foo":"bar"}'
      ```

8. Go to your `receiver` folder and run the `init` Kyma CLI command to to initialise the scaffold for your second Function:
   ```bash
   kyma init function
   ```
9.  The `init` command creates the same files as in the `emitter` folder.
10. In the `config.yaml` file configure which event types your Function should subscribe to:
    ```yaml
    name: event-receiver
    namespace: default
    runtime: nodejs14
    source:
       sourceType: inline
    subscriptions:
       - name: event-receiver
         protocol: ""
         filer:
            filters:
               - eventSource:
                   property: source
                   type: exact
                   value: ""
                eventType:
                   property: type
                   type: exact
                   value: sap.kyma.custom.acme.payload.sanitised.v1
    ```
11.  Apply your receiver Function:

   ```bash
   kyma apply function
   ```
   The Function is configured, built and deployed in Kyma runtime. The Susbscription becomes active and all events with selected type will be processed by the Function.  

12.  Test the whole setup  
Send a payload to the first Function (for example using the above-mentioned POST request). As the Functions are joined by the In-cluster Eventing, the payload is directed to both of your Functions.
You can see (for example in the Function logs) that both, sanitization logic (using the first Function) and the storing logic (using the second Function) are executed.

Summary:

You can distribute application logic into specialised, event-driven Functions