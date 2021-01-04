---
title: Trigger a Function with an event
type: Tutorials
---

This tutorial shows how to trigger a Function with an event from an Application connected to Kyma.

> **NOTE:** To learn more about events flow in Kyma, read the [eventing](/components/event-mesh) documentation.

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](#tutorials-create-a-function) tutorial.

You must also have:

- An Application bound to a specific Namespace. Read the tutorials to learn how to [create](/components/application-connector#tutorials-create-a-new-application) an Application and [bind](/components/application-connector#tutorials-bind-an-application-to-a-namespace) it to a Namespace.
- An event service (an API of [AsyncAPI](https://www.asyncapi.com/) type) registered in the desired Application. Read the [tutorial](/components/application-connector#tutorials-register-a-service) to learn how to do it.
- A Service Instance created for the registered service to expose events in a specific Namespace. Read the [tutorial](/components/application-connector#tutorials-bind-a-service-to-a-namespace) for details.

## Steps

Follows these steps:

<div tabs name="steps" group="trigger-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Run the `kubectl get brokers -n {NAMESPACE}` command to check if there already is the Knative's `default` Broker running in the Namespace where your Function is running. If not, you must manually inject the Broker into the Namespace to enable Trigger creation and event flow. To do that, run this command:

  ```bash
  kubectl label namespace {NAMESPACE} knative-eventing-injection=enabled
  ```

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export APP_NAME={APPLICATION_NAME}
    export EVENT_VERSION={EVENT_TYPE_VERSION}
    export EVENT_TYPE={EVENT_TYPE_NAME}
    ```

    > **NOTE:** Function takes the name from the Function CR name. The Trigger CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

These variables refer to the following:

- **APP_NAME** is the name of the Application CR which is the source of the events.
- **EVENT_VERSION** points to the specific event version type, such as `v1`.
- **EVENT_TYPE** points to the event type to which you want to subscribe your Function, such as `user.created`.

3. Create a Trigger CR for your Function to subscribe your Function to a specific event type.

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
          apiVersion: v1
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

2. In the left navigation panel, go to **Workloads** > **Functions** and navigate to your Function.

3. Once in the Function details view, Switch to the **Configuration** tab, and select **Add Event Trigger** in the **Event Triggers** section.

4. Select the event type and version that you want to use as a trigger for your Function and select **Add** to confirm changes.

The message appears on the UI confirming that the Event Trigger was successfully created, and you will see it in the **Event Triggers** section in your Function.

  </details>
</div>

## Test the trigger

> **CAUTION:** Before you follow steps in this section and send a sample event, bear in mind that it will be propagated to all services subscribed to this event type.

To test if the Trigger CR is properly connected to the Function:

1. Change the Function's code to:​

    ```js
    module.exports = {
      main: function (event, context) {
        console.log("User created: ", event.data);
      }
    }
    ```

2.  Send an event manually to trigger the function. The first example shows the implementation introduced with the Kyma 1.11 release where a [CloudEvent](https://github.com/cloudevents/spec/blob/v1.0/spec.md) is sent directly to the Event Mesh. In the second example, an event also reaches the Event Mesh, but it is first modified by the compatibility layer to the format compliant with the CloudEvents specification. This solution ensures compatibility if your events follow a format other than CloudEvents, or you use the Event Bus available before 1.11.

    > **TIP:** For details on CloudEvents, exposed endpoints, and the compatibility layer, read about [event processing and delivery](/components/event-mesh/#details-event-processing-and-delivery).

    <div tabs name="examples" group="test=trigger">
      <details>
      <summary label="CloudEvents">
      Send CloudEvents directly to Event Mesh
      </summary>

    ```bash
    curl -v -H "Content-Type: application/cloudevents+json" https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/events -k --cert {CERT_FILE_NAME} --key {KEY_FILE_NAME} -d \
      '{
        "ce-specversion": "1.0",
        "ce-source": "{APP_NAME}",
        "ce-type": "{EVENT_TYPE}",
        "ce-eventtypeversion": "{EVENT_VERSION}",
        "ce-id": "A234-1234-1234",
        "data": "123456789",
        "datacontenttype": "application/json"
      }'
    ```
      </details>
      <details>
      <summary label="Compatibility layer">
      Send events to Event Mesh through compatibility layer
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

    - **CERT_FILE_NAME** and **KEY_FILE_NAME** are client certificates for a given Application. You can get them by completing steps in the [tutorial](/components/application-connector/#tutorials-get-the-client-certificate).

3. After sending an event, you should get this result from logs of your Function's latest Pod:

    ```text
    User created: 123456789
    ```
