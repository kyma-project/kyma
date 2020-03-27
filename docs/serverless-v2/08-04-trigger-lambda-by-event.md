---
title: Trigger a lambda with events
type: Tutorials
---

This tutorial shows how to trigger a lambda with an event from a specific application connected to Kyma.

>**NOTE:** To learn more about events flow in Kyma, read [this](/components/knative-eventing-mesh) topic.

## Prerequisites

This tutorial is based on an existing lambda. To create one, follow the [Create a lambda](#tutorials-create-a-lambda) tutorial.

Also you must have: 

- Created Application and bounded it to a specific Namespace. Learn how to [create](/components/application-connector#tutorials-create-a-new-application) an Application and [bind](/components/application-connector#tutorials-bind-an-application-to-a-namespace) an Application to a Namespace.
- Registered a service with events in the desired Application (with [AsyncAPI](https://www.asyncapi.com/) specification). Learn how to do [it](components/application-connector/#tutorials-register-a-service).
- Created a Service Instance for the registered service to expose events in specific namespace. See [this](tu bedzie dokument) for instruction.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    export APP_NAME={APP_NAME}
    ```

    > **NOTE:** Lambda takes the name from the Function CR name. The Trigger CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

    > **NOTE:** **APP_NAME** is taken from the Application name (more precisely from Application CR name).

2. Create an Trigger CR for your lambda. It is subscribed your lambda to specific event.

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

    - The **spec.filter.attributes.eventtypeversion** points to the version of event (for example `v1`) and **spec.filter.attributes.type** points to type of event (for example `user.created`).

## Test the trigger

To test if the Trigger has been properly connected to the lambda:

1. Change the lambda's code to something like that:​

    ```js
    module.exports = {
      main: function (event, context) {
        console.log("User created: ", event.data);
      }
    }
    ```

2. Send an event to trigger the lambda. Below is the way to send an event manually.

    ```bash
    curl -X POST https://gateway.{CLUSTER_DOMAIN}/$APP_NAME/v1/events -k --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -d \
    '{
        "event-type": "user.created",
        "event-type-version": "v1",
        "event-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
        "event-time": "2020-04-02T21:37:00Z",
        "data": "123456789"
    }'
    ```

    - **CLUSTER_DOMAIN** is domain of your cluster. For example `kyma.local`.

    - **CERT_FILE_NAME** and **KEY_FILE_NAME** are a credentials of client certificates for a given Application. You can get they from [this](https://kyma-project.io/docs/master/components/application-connector/#tutorials-get-the-client-certificate) tutorial.

3. After sending event, you should get this result in logs of pod of your lambda:

    ```text
    User created: 123456789
    ```



