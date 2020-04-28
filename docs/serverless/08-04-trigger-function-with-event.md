---
title: Trigger a function with an event
type: Tutorials
---

This tutorial shows how to trigger a function with an event from an Application connected to Kyma.

> **NOTE:** To learn more about events flow in Kyma, read the [eventing](/components/knative-eventing-mesh) documentation.

## Prerequisites

This tutorial is based on an existing function. To create one, follow the [Create a function](#tutorials-create-a-function) tutorial.

You must also have:

- An Application bound to a specific Namespace. Read the tutorials to learn how to [create](/components/application-connector#tutorials-create-a-new-application) an Application and [bind](/components/application-connector#tutorials-bind-an-application-to-a-namespace) it to a Namespace.
- An event service (an API of [AsyncAPI](https://www.asyncapi.com/) type) registered in the desired Application. Learn [here](components/application-connector/#tutorials-register-a-service) how to do it.
- A Service Instance created for the registered service to expose events in a specific Namespace. See [this](/components/application-connector/#tutorials-bind-a-service-to-a-namespace) tutorial for details.

## Steps

Follows these steps:

<div tabs name="steps" group="trigger-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export APP_NAME={APPLICATION_NAME}
    export EVENT_VERSION={EVENT_TYPE_VERSION}
    export EVENT_TYPE={EVENT_TYPE_NAME}
    ```

    > **NOTE:** Function takes the name from the Function CR name. The Trigger CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

These variables refer to the following:

- **APP_NAME** is taken from the name of the Application CR and specifies the source of events.
- **EVENT_VERSION** points to the specific event version, such as `v1`.
- **EVENT_TYPE** points to the given event type to which you want to subscribe your function, such as `user.created`.

2. Create a Trigger CR for your function to subscribe your function to a specific event type.

    ```yaml
    cat <<EOF | kubectl apply -f  -
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Trigger
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      broker: default
      filter:
        attributes:
          eventtypeversion: $EVENT_VERSION
          source: $APP_NAME
          type: $EVENT_TYPE
      subscriber:
        ref:
          apiVersion: serving.knative.dev/v1
          kind: Service
          name: $NAME
          namespace: $NAMESPACE
    EOF
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. From the drop-down list in the top navigation panel, select the Namespace in which your Application exposes events.

2. Go to the **Functions** view in the left navigation panel and navigate to your function.

3. Once in the function view, Switch to the **Configuration** tab, and select **Add Event Trigger** in the **Event Triggers** section.

4. Select the event type and version that you want to use as a trigger for your function and select **Add** to confirm changes.

The message appears on the UI confirming that the Event Trigger was successfully created, and you will see it in the **Event Triggers** section in your function.

    </details>
</div>

## Test the trigger

> **CAUTION:** Before you follow steps in this section and send a sample event, bear in mind that it will be propagated to all services subscribed to this event type.

To test if the Trigger CR is properly connected to the function:

1. Change the function's code to:​

    ```js
    module.exports = {
      main: function (event, context) {
        console.log("User created: ", event.data);
      }
    }
    ```

2.  Send an event manually to trigger the function. In the first example, the payload complies with the [CloudEvents](https://github.com/cloudevents/spec/blob/v1.0/spec.md) specification and the event is sent directly to Eventing Mesh. If you want to send the events to compatibility layer which forwards them to Eventing Mesh, use the second example. 

    <div tabs name="examples" group="test=trigger">
      <details>
      <summary label="CloudEvents">
      CloudEvents
      </summary>

    ```bash
       curl -v -H "Content-Type: application/cloudevents+json" https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/events -k --cert {CERT_FILE_NAME} --key {KEY_FILE_NAME} -d \
          '{
            "specversion": "1.0",
            "source": "{APP_NAME}",
            "type": "{EVENT_TYPE}",
            "eventtypeversion": "{EVENT_VERSION}",
            "id": "A234-1234-1234",
            "data": "123456789",
            "datacontenttype": "application/json"
          }' 
    ```
      </details>
      <details>
      <summary label="Compatibility layer">
      Compatibility layer
      </summary>

    ```bash
      curl -H "Content-Type: application/json" https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events -k --cert {CERT_FILE_NAME} --key {KEY_FILE_NAME} -d \
        '{
            "event-type": "{EVENT_TYPE}",
            "event-type-version": "{EVENT_VERSION}",
            "event-time": "2020-04-02T21:37:00Z",
            "data": "123456789"
         }'
    ``` 

      </details>
  </div>

    - **CLUSTER_DOMAIN** is the domain of your cluster, such as `kyma.local`.

    - **CERT_FILE_NAME** and **KEY_FILE_NAME** are client certificates for a given Application. You can get them by completing steps in [this](/components/application-connector/#tutorials-get-the-client-certificate) tutorial.

3. After sending an event, you should get this result from logs of your function's latest Pod:

    ```text
    User created: 123456789
    ```



