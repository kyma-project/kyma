# Send and Receive Cloud Events

This tutorial demonstrates how to connect two Functions asynchronously with [cloud events](https://github.com/cloudevents/spec). It is based on the [in-cluster Eventing example](https://github.com/kyma-project/serverless/tree/main/examples/incluster_eventing).

The example provides a very simple scenario of asynchronous communication between two Functions. The first Function accepts the incoming traffic using HTTP, sanitizes the payload, and publishes the content as an in-cluster cloud event using [Kyma Eventing](https://kyma-project.io/docs/kyma/latest/01-overview/eventing/).
The second Function is a message receiver. It subscribes to the given event type and stores the payload.

This tutorial shows only one possible use case. There are many more use cases on how to orchestrate your application logic into specialized Functions and benefit from decoupled, re-usable components and event-driven architecture.

## Prerequisites

- [Kyma CLI](https://github.com/kyma-project/cli)
- [Nats, Eventing, Istio, and API-Gateway modules added](https://kyma-project.io/#/02-get-started/01-quick-install)

## Steps

1. Export the `KUBECONFIG` variable:

   ```bash
   export KUBECONFIG={KUBECONFIG_PATH}
   ```
2. Enable Istio service mesh for `default` namespace:

   ```bash
   kubectl label namespaces default istio-injection=enabled
   ```

3. Create the `emitter` and `receiver` folders in your project.

### Create the Emitter Function

1. Go to the `emitter` folder and run Kyma CLI `init` command to initialize the scaffold for your first Function:

   ```bash
   kyma alpha function init
   ```

   The `init` command creates these files in your workspace folder:

   - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
   - `package.json` with the Function's dependencies

2. Provide your Function logic in the `handler.js` file:

   > [!NOTE]
   > In this example, there's no real sanitization logic but the Function simply logs the payload.

   ```js
   const { SpanStatusCode } = require("@opentelemetry/api");

   module.exports = {
      main: async function (event, context) {
         let sanitisedData = sanitise(event.data)

         const eventType = "payload.sanitised";
         const eventSource = "my-app";

         const span = event.tracer.startSpan('call-to-kyma-eventing');
         
         // you can pass additional cloudevents attributes  
         // const eventtypeversion = "v1";
         // const datacontenttype = "application/json";
         // return await event.emitCloudEvent(eventType, eventSource, sanitisedData, {eventtypeversion, datacontenttype})
         
         return await event.emitCloudEvent(eventType, eventSource, sanitisedData)
               .then(resp => {
                  console.log(resp.status);
                  span.addEvent("Event sent");
                  span.setAttribute("event-type", eventType);
                  span.setAttribute("event-source", eventSource);
                  span.setStatus({code: SpanStatusCode.OK});
                  return "Event sent";
               }).catch(err=> {
                  console.error(err)
                  span.setStatus({
                     code: SpanStatusCode.ERROR,
                     message: err.message,
                  });
                  return err.message;
               }).finally(()=>{
                  span.end();
               });
      }
   }
   let sanitise = (data)=>{
      console.log(`sanitising data...`)
      console.log(data)
      return data
   }
   ```

   Include opentelemetry SDK in the Function dependencies. Add the following to the `package.json`:
   ```js
   {
      "dependencies": {
         "@opentelemetry/api": "^1.0.4"
      }
   }
   ```


   The `payload.sanitised` is a sample event type that the emitter Function uses when publishing events. You can choose a different one that better suits your use case. Keep in mind the constraints described on the [Event names](https://kyma-project.io/docs/kyma/latest/05-technical-reference/evnt-01-event-names/) page. The receiver subscribes to the event type to consume the events.

   The `event` object provides a convenient API for emitting events. To learn more, read [Function's specification](../technical-reference/07-70-function-specification.md#event-object-sdk).
   
3. Apply your emitter Function:

   ```bash
   kyma alpha function create emitter --source handler.js --dependencies package.json
   ```

   Your Function is now built and deployed in Kyma runtime. Kyma exposes it through the APIRule. The incoming payloads are processed by your emitter Function. It then sends the sanitized content to the workload that subscribes to the selected event type. In our case, it's the receiver Function.

4. Expose Function by creating the APIRule CR:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v2alpha1
   kind: APIRule
   metadata:
     name: incoming-http-trigger
   spec:
     hosts:
     - incoming
     service:
       name: emitter
       namespace: default
       port: 80
     gateway: kyma-system/kyma-gateway
     rules:
     - path: /*
       methods: ["GET", "POST"]
       noAuth: true
   EOF
   ```
5. Run the following command to get the domain name of your Kyma cluster:

    ```bash
    kubectl get gateway -n kyma-system kyma-gateway \
        -o jsonpath='{.spec.servers[0].hosts[0]}'
    ```

6. Export the result without the leading `*.` as an environment variable:

    ```bash
    export DOMAIN={DOMAIN_NAME}

7. Test the first Function. Send the payload and see if your HTTP traffic is accepted:

   ```bash
   curl -X POST "https://incoming.${DOMAIN}" -H 'Content-Type: application/json' -d '{"foo":"bar"}'
   ```
   
   You should see the `Event sent` message as a response.

### Create the Receiver Function

1. Go to your `receiver` folder and run Kyma CLI `init` command to initialize the scaffold for your second Function:

   ```bash
   kyma alpha function init
   ```

   The `init` command creates the same files as in the `emitter` folder.
   In the following example, the receiver function logs the received payload.

   ```js
   module.exports = {
      main: function (event, context) {
         store(event.data)
         return 'OK'
      }
   }
   let store = (data)=>{
      console.log(`storing data...`)
      console.log(data)
      return data
   }
   ```

3. Apply your receiver Function:

   ```bash
   kyma alpha function create receiver --source handler.js --dependencies package.json
   ```

   The Function is configured, built, and deployed in Kyma runtime. The Subscription becomes active and all events with the selected type are processed by the Function.  

2. Subscribe the `receiver` Function to the event:  

   ```bash
   cat <<EOF | kubectl apply -f -
      apiVersion: eventing.kyma-project.io/v1alpha2
      kind: Subscription
      metadata:
         name: event-receiver
         namespace: default
      spec:
         sink: 'http://receiver.default.svc.cluster.local'
         source: "my-app"
         types:
         - payload.sanitised
   EOF
   ```

### Test the Whole Setup

Send a payload to the first Function. For example, use the POST request mentioned above. As the Functions are joined by the in-cluster Eventing, the payload is processed in sequence by both of your Functions.
In the Function's logs, you can see that both sanitization logic (using the first Function) and the storing logic (using the second Function) are executed.
