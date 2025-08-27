# Kyma Eventing - Basic Diagnostics

## Symptom

- You publish an event but the event is not received by the subscriber.

- Subscription is not in `Ready` state.

- You are unable to publish events.

## Cause

Trouble with Kyma Eventing may be caused by various issues, so this document guides you through the diagnostic steps to determine the specific root cause.

## Solution

Follow these steps to detect the source of the problem:

### Step 1: Check the Status of the Eventing Custom Resource (CR)

1. Check the Eventing CR. Is the **State** field `Ready`?

    ```bash
    kubectl -n kyma-system get eventings.operator.kyma-project.io
    ```

2. If **State** is not `Ready`, check the exact reason of the error in the status of the CR by running the command:

    ```bash
    kubectl -n kyma-system get eventings.operator.kyma-project.io eventing -o yaml
    ```

3. If the **State** is `Ready`, the Eventing CR is not an issue. Follow the next steps to find the source of the problem.

### Step 2: Check the Status of the Subscription

1. Check whether the Subscription is `Ready`. Run the command:

    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME}
    ```

2. If the Subscription is not `Ready`, check the exact reason of the error in the status of the Subscription by running the command:

    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o yaml
    ```

    If the status of the Subscription informs you that the sink is not a valid cluster local svc, investigate the Subscription sink. Verify if the sink is a valid HTTP endpoint, for example: `test.test.svc.cluster.local`.

3. Check if the event type defined in the Subscription is correctly formatted as specified in the [Event names](../evnt-event-names.md) guidelines.

### Step 3: Check if the Event Was Published Correctly

1. Check the HTTP status code returned after sending an event.

   - If the HTTP status code is 4xx, check if you are sending the events in correct formats. Eventing supports two event formats (legacy and CloudEvents); see [Eventing tutorials](../tutorials/evnt-01-prerequisites.md) for more information.
   - If the HTTP status code is 5xx, check the logs from the Eventing publisher proxy Pod for any errors. To fetch these logs, run this command:

      ```bash
      kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=eventing-publisher-proxy
      ```

   - If the HTTP status code is 2xx but it's still not working, verify if you are sending the event on the same event type as defined in the Subscription.

      ```bash
      kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o jsonpath='{.spec.filter.filters}'
      ```

### Step 4: Check the Eventing Manager Logs

1. Check the logs from the Eventing Manager Pod for any errors and to verify that the event is dispatched.
   To fetch these logs, run this command:

    ```bash
    kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing-manager,app.kubernetes.io/name=eventing-manager
    ```

2. Check for any error messages in the logs. If the event dispatch log `"message":"event dispatched"` is not present for NATS backend, the issue could be one of the following:

   - The subscriber (the sink) is not reachable or the subscriber cannot process the event. Check the logs of the subscriber instance.

   - The event was published in a wrong format.

   - Eventing Manager cannot connect to NATS Server.

### Step 5: Check if the Subscription Sink Is Healthy

1. Check whether the workload URL defined in the Subscription sink is correct and healthy to receive events. To get the sink from the Subscription, run this command:

    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o jsonpath='{.spec.sink}'
    ```

2. To check the health of the sink, run the following commands:

    ```bash
    kubectl -n default run --image=curlimages/curl --restart=Never sink-test-tmp -- curl --head {SINK_URL}
    kubectl -n default logs sink-test-tmp 
    kubectl -n default delete pod sink-test-tmp
    ```

    If the returned HTTP status code is not 2xx, check the logs of the subscriber instance.

### Step 6: Check NATS JetStream Status

1. Check the health of NATS Pods. Run the command:

    ```bash
    kubectl -n kyma-system get pods -l nats_cluster=eventing-nats
    ```

2. Check if the stream and consumers exist in NATS JetStream by following the [JetStream troubleshooting guide](evnt-02-jetstream-troubleshooting.md).

If you can't find a solution, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to our [Slack channel](https://kyma-community.slack.com/) to get direct support from the community.
