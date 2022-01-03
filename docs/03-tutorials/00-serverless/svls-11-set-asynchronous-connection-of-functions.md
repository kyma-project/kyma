---
title: Set asynchronous communication between Functions
---

This tutorial shows how you can connect two Functions using asynchronous communication. You can do it thanks to the In-cluster Eventing embedded in Kyma. The tutorial is based on Function from [Incluster Eventing example](https://github.com/kyma-project/examples/pull/188).

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
3. Go to your `emitter` folder and run the `init` Kyma CLI command to initiate your Function:

   ```bash
   kyma init function
    ```

  The `init` command creates these files in your workspace folder:

  - `config.yaml`	with the Function's configuration

>**NOTE:** See the detailed description of all fields available in the [`config.yaml` file](../../05-technical-reference/svls-06-function-configuration-file.md).

  - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
  - `package.json` with the Function's dependencies

4. In the `config.yaml` file configure your API Rule to allow for the incoming outside traffic (using HTTP). Enter your cluster domain in the host key:

  ```yaml
  apiRules:
    - name: incomming-http-trigger
      service:
        host: incomming.inclusterevent.acme.shoot.canary.k8s-hana.ondemand.com
      rules:
        - methods:
            - GET
          accessStrategies:
            - handler: allow
  ```

5. Provide your Function logic in the `handler.js` file. In the example you can find a simple code sanitization:

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

6. Apply your emitter Function:

  ```bash
  kyma apply function
  ```
Your emitter Function sends the event, altered in a required way, to the receiver Function.

7. To send the event you need to build the Cloud Event. To learn more please visit the [Function's Specification](https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-08-function-specification#event-object-sdk) page.
In addition, your `eventOut.source` key needs to point at `“kyma”` to use Kyma In-cluster Eventing.

8. Go to your `receiver` folder and run the `init` Kyma CLI command to initiate your function:
   ```bash
   kyma init function
   ```
9. The `init` command creates the same files as in the emitter folder.
10. In the `config.yaml` file configure your Subscriptions to allow for receipt and storage of the event data:
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
11. The `value` key in your `handler.js` file must consist of seven elements separated by dots, for example `sap.kyma.custom.acme.payload.sanitised.v1`. First three elements must remain unchanged (`sap.kyma.custom`). The last element specifies the version of the event.
12. When you call your emitter Function, your receiver Function must also respond.