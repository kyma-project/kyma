---
title: Set asynchronous communication between Functions
---

This tutorial demonstrates how to connect two Functions asynchronously. It is based on the [in-cluster Eventing example](https://github.com/kyma-project/examples/tree/main/incluster_eventing).

The example provides a very simple scenario of asynchronous communication between two Functions. The first Function accepts the incoming traffic via HTTP, sanitizes the payload, and publishes the content as an in-cluster event using [Kyma Eventing](../../01-overview/eventing).
The second Function is a message receiver. It subscribes to the given event type and stores the payload.

This tutorial shows only one possible use case. There are many more use cases on how to orchestrate your application logic into specialized Functions and benefit from decoupled, re-usable components and event-driven architecture.

## Prerequisites


- [Kyma CLI](https://github.com/kyma-project/cli)
- [Kyma installed](../../04-operation-guides/operations/02-install-kyma.md) locally or on a cluster
- [Istio sidecar injection enabled](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md) in the Namespace in which you want to deploy the Functions
## Steps

1. Export the `KUBECONFIG` variable:
   ```bash
   export KUBECONFIG={KUBECONFIG_PATH}
   ```
2. Create the `emitter` and `receiver` folders in your project.

### Create the emitter Function

1. Go to the `emitter` folder and run Kyma CLI `init` command to initialize the scaffold for your first Function:

   ```bash
   kyma init function
    ```

  The `init` command creates these files in your workspace folder:

  - `config.yaml`	with the Function's configuration

      >**NOTE:** See the detailed description of all fields available in the [`config.yaml` file](../../05-technical-reference/svls-06-function-configuration-file.md).

  - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
  - `package.json` with the Function's dependencies

2. In the `config.yaml` file, configure an APIRule to expose your Function to the incoming traffic over HTTP. Provide the subdomain name in the `host` property:

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

3. Provide your Function logic in the `handler.js` file:
>**NOTE:** In this example, there's no sanitization logic. The `sanitize` Function is just a placeholder.

   ```js
   module.exports = {
      main: async function (event, context) {
         let sanitisedData = sanitise(event.data)

         const eventType = "sap.kyma.custom.acme.payload.sanitised.v1";
         const eventSource = "kyma";
         
         return await event.emitCloudEvent(eventType, eventSource, sanitisedData)
               .then(resp => {
                  return "Event sent";
               }).catch(err=> {
                  console.error(err)
                  return err;
               });
      }
   }
   let sanitise = (data)=>{
      console.log(`sanitising data...`)
      console.log(data)
      return data
   }
   ```
   >**NOTE:** The `sap.kyma.custom.acme.payload.sanitised.v1` is a sample event type declared by the emitter Function when publishing events. You can choose a different one that better suits your use case. Keep in mind the constraints described on the [Event names](../../05-technical-reference/evnt-01-event-names.md) page. The receiver subscribes to the event type to consume the events.

   >**NOTE:** The event object provides convenience functions to build and publish events. To send the event, build the Cloud Event. To learn more, read [Function's specification](../../05-technical-reference/svls-07-function-specification.md#event-object-sdk). In addition, your **eventOut.source** key must point to `“kyma”` to use Kyma in-cluster Eventing.
   >**NOTE:** There is a `require('axios')` line even though the Function code is not using it directly. This is needed for the auto-instrumentation to properly handle the outgoing requests sent using the `publishCloudEvent` method (which uses `axios` library under the hood). Without the `axios` import the Function still works, but the published events are not reflected in the trace backend.

4. Apply your emitter Function:

    ```bash
    kyma apply function
    ```
   Your Function is now built and deployed in Kyma runtime. Kyma exposes it through the APIRule. The incoming payloads are processed by your emitter Function. It then sends the sanitized content to the workload that subscribes to the selected event type. In our case, it's the receiver Function.

5. Test the first Function. Send the payload and see if your HTTP traffic is accepted:

      ```bash
      export KYMA_DOMAIN={KYMA_DOMAIN_VARIABLE}
   
      curl -X POST https://incoming.${KYMA_DOMAIN} -H 'Content-Type: application/json' -d '{"foo":"bar"}'
      ```
### Create the receiver Function

1. Go to your `receiver` folder and run Kyma CLI `init` command to initialize the scaffold for your second Function:
   ```bash
   kyma init function
   ```
   The `init` command creates the same files as in the `emitter` folder.

2. In the `config.yaml` file, configure event types your Function will subscribe to:

<div tabs name="function" group="set-asynchronous-connection-of-functions">
  <details>
  <summary label="v1alpha1">
  v1alpha1
  </summary>
   
   ```yaml
    name: event-receiver
    namespace: default
    runtime: nodejs18
    source:
       sourceType: inline
    subscriptions:
       - name: event-receiver
         protocol: ""
         filter:
            filters:
               - eventSource:
                   property: source
                   type: exact
                   value: ""
                eventType:
                   property: type
                   type: exact
                   value: sap.kyma.custom.acme.payload.sanitised.v1
    schemaVersion: v0
   ```

</details>
<details>
  <summary label="v1alpha2">
  v1alpha2
  </summary>   

```yaml
    name: event-receiver
    namespace: default
    runtime: nodejs18
    source:
       sourceType: inline
    subscriptions:
       - name: event-receiver
         typeMatching: exact
         source: ""
         types:
           - sap.kyma.custom.acme.payload.sanitised.v1
    schemaVersion: v1
   ```

</details>
</div>

3.  Apply your receiver Function:
     ```bash
     kyma apply function
     ```
   The Function is configured, built, and deployed in Kyma runtime. The Subscription becomes active and all events with the selected type are processed by the Function.  

### Test the whole setup  
Send a payload to the first Function. For example, use the POST request mentioned above. As the Functions are joined by the in-cluster Eventing, the payload is processed in sequence by both of your Functions.
In the Function's logs, you can see that both sanitization logic (using the first Function) and the storing logic (using the second Function) are executed.
