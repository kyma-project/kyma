---
title: Set asynchronous communication between Functions
---

This tutorial demonstrates how to connect two Functions asynchronously. It is based on the [in-cluster Eventing example](https://github.com/kyma-project/examples/tree/main/incluster_eventing).

The example provides a very simple scenario of asynchronous communication between two Functions. The first Function accepts the incoming traffic via HTTP, sanitizes the payload, and publishes the content as an in-cluster event via [Kyma Eventing](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/eventing/).
The second Function is a message receiver. It subscribes to the given event type and stores the payload.

## Prerequisites


- [Kyma CLI](https://github.com/kyma-project/cli)
- Kyma installed locally or on a cluster

## Initial steps

1. Export the `KUBECONFIG` variable:
   ```bash
   export KUBECONFIG={KUBECONFIG_PATH}
   ```
2. Create two folders in your project - `emitter` and `receiver`.

## Create the emitter Function

1. Go to the `emitter` folder and run Kyma CLI `init` command to initialize a scaffold for your first Function:

   ```bash
   kyma init function
    ```

  The `init` command creates these files in your workspace folder:

  - `config.yaml`	with the Function's configuration

   >**NOTE:** See the detailed description of all fields available in the [`config.yaml` file](../../05-technical-reference/svls-06-function-configuration-file.md).

  - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
  - `package.json` with the Function's dependencies

2. In the `config.yaml` file, configure an API Rule to expose your Function to the incoming traffic over HTTP. Provide the subdomain name in the `host` property:

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

3. Provide your Function logic in the `handler.js` file. In the following example, there's no actual sanitization logic, `sanitise` Function is just a placeholder:

   ```js
   const { v4: uuidv4 } = require('uuid');
   module.exports = {
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
   >**NOTE:** The `sap.kyma.custom.acme.payload.sanitised.v1` is a sample event type declared by the emitter Function when publishing events. You can choose a different one that better suits your use case. Keep in mind the constraints described on the [Event names](https://kyma-project.io/docs/kyma/latest/05-technical-reference/evnt-01-event-names) page. The receiver subscribes to the event type to consume the events.

   >**NOTE:** The event object provides convenience functions to build and publish events. To send the event, build the Cloud Event. To learn more, read [Function's specification](https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-08-function-specification#event-object-sdk). In addition, your **eventOut.source** key must point to `“kyma”` to use Kyma in-cluster Eventing.

4. Apply your emitter Function:

    ```bash
    kyma apply function
    ```
   Your Function is now built and deployed in Kyma runtime. Kyma will expose it via API Rule. Any incoming payload will be processed by your emitter Function. It sends the sanitized content to whatever workload that subscribes to the selected event type - in our case - the receiver Function.

5. Test the first Function. Send a paylod and see if your HTTP traffic is accepted:

      ```bash
      export KYMA_DOMAIN=.... # export the variable for you Kyma domain
      
      curl -X POST https://incoming.${KYMA_DOMAIN}
      -H 'Content-Type: application/json'
      -d '{"foo":"bar"}'
      ```
## Create the receiver Function

1. Go to your `receiver` folder and run the `init` Kyma CLI command to to initialise the scaffold for your second Function:
   ```bash
   kyma init function
   ```
2.  The `init` command creates the same files as in the `emitter` folder.
3. In the `config.yaml` file configure which event types your Function should subscribe to:
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
4.  Apply your receiver Function:
     ```bash
     kyma apply function
     ```
   The Function is configured, built and deployed in Kyma runtime. The Susbscription becomes active and all events with selected type will be processed by the Function.  

5.  Test the whole setup  
Send a payload to the first Function (for example using the above-mentioned POST request). As the Functions are joined by the In-cluster Eventing, the payload is processed in sequence by both of your Functions.
You can see (for example in the Function logs) that both, sanitization logic (using the first Function) and the storing logic (using the second Function) are executed.

## Summary

This tutorial demonstrated only one possible use case. You can find plenty of use cases yourself on how to orchestrate your application logic into specialised Functions and benefit from decoupled, re-usable components and event-driven architecture.