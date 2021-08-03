---
title: Recive and publish a CloudEvent
---

This tutorial shows how you can receive a [Cloud Events](https://cloudevents.io/) in the Function and publish a CloudEvent to the same queue using Functions SDK.

1. [Create a Function](./svls-02-create-git-function.md) then modify its body to be ready to publish a CloudEvent and replace `{event_type}` with the desired event type:

    ```js
        module.exports = {
            main: function (event, context) {
                console.log("publish an event");
                // Alternatively, you can build a Cloud Event object manually to get more control over object's fields:
                // let ce = {
                //   'eventtypeversion': {EVENT_EVENTTYPEVERSION},
                //   'specversion': {EVENT_SPECVERSION},
                //   'source': {EVENT_SOURCE},
                //   'data': {EVENT_DATA},
                //   'type': {EVENT_TYPE},
                //   'id': {EVENT_ID},
                //  };
                let ce = event.buildResponseCloudEvent(
                "A234-4321-4321",
                "{EVENT_TYPE}",
                "sample event data"
                );

                event.publishCloudEvent(ce);
                console.log("done");
            }
        }
    ```

2. Create a Subscription custom resource (CR) to subscribe the Function to the expected event type:

    ```bash
    cat <<EOF | kubectl apply -f  -
        apiVersion: eventing.kyma-project.io/v1alpha1
        kind: Subscription
        metadata:
        name: {SUBSCRIPTION_NAME}
        namespace: {SUBSCRIPTION_NAMESPACE}
        spec:
        filter:
            filters:
            - eventSource:
                property: source
                type: exact
                value: ""
            eventType:
                property: type
                type: exact
                value: {EVENT_TYPE}
        protocol: ""
        protocolsettings: {}
        sink: http://{FUNCTION_NAME}.{FUNCTION_NAMESPACE}.svc.cluster.local
    EOF
    ```

3. Now you can send a CloudEvent to the Function and Function will publish a new CloudEvent based on the first one.
