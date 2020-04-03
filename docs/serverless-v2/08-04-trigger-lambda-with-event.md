---
title: Trigger a lambda with an event
type: Tutorials
---

This tutorial shows how to trigger a lambda with an event from an Application connected to Kyma.

> **NOTE:** To learn more about events flow in Kyma, read the [eventing](/components/knative-eventing-mesh) documentation.

## Prerequisites

This tutorial is based on an existing lambda. To create one, follow the [Create a lambda](#tutorials-create-a-lambda) tutorial.

You must also have: 

- An Application bound to a specific Namespace. Read the tutorials to learn how to [create](/components/application-connector#tutorials-create-a-new-application) an Application and [bind](/components/application-connector#tutorials-bind-an-application-to-a-namespace) it to a Namespace.
- An event service (an API of [AsyncAPI](https://www.asyncapi.com/) type) registered in the desired Application. Learn [here](components/application-connector/#tutorials-register-a-service) how to do it.
- A Service Instance created for the registered service to expose events in a specific Namespace. See [this](/components/application-connector/#tutorials-bind-a-service-to-a-namespace) tutorial for details.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    export APP_NAME={APP_NAME}
    ```

    > **NOTE:** Lambda takes the name from the Function CR name. The Trigger CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

    > **NOTE:** **APP_NAME** is taken from the name of the Application CR.

2. Create a Trigger CR for your lambda to subscribe your lambda to a specific event type.

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
          eventtypeversion: {EVENT_VESRION}
          source: {APP_NAME}
          type: {EVENT_TYPE}
      subscriber:
        ref:
          apiVersion: serving.knative.dev/v1
          kind: Service
          name: $NAME
          namespace: $NAMESPACE
    EOF
    ```

    The **spec.filter.attributes.eventtypeversion** parameter points to the specific event version, such as `v1`, and **spec.filter.attributes.type** points to the given event type, such as `user.created`.

## Trigger the lambda

To test if the Trigger CR is properly connected to the lambda:

1. Change the lambda's code to:​

    ```js
    module.exports = {
      main: function (event, context) {
        console.log("User created: ", event.data);
      }
    }
    ```

2. Send an event manually to trigger the lambda:

    ```bash
    curl -X POST https://gateway.{CLUSTER_DOMAIN}/$APP_NAME/v1/events -k --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -d \
    '{
        "event-type": "{EVENT_TYPE}",
        "event-type-version": "{EVENT_VESRION}",
        "event-time": "2020-04-02T21:37:00Z",
        "data": "123456789"
    }'
    ```

    - **CLUSTER_DOMAIN** is the domain of your cluster, such as `kyma.local`.

    - **CERT_FILE_NAME** and **KEY_FILE_NAME** are client certificates for a given Application. You can get them by completing steps in [this](/components/application-connector/#tutorials-get-the-client-certificate) tutorial.

3. After sending an event, you should get this result from logs of your lambda's Pod:

    ```text
    User created: 123456789
    ```
