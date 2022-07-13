---
title: Kyma Eventing - Basic Diagnostics
---

## Symptom

- You publish an event but the event is not received by the subscriber.

- Subscription is not in ready state.

- You are unable to publish events.

## Cause

Trouble with Kyma Eventing may be caused by various issues, so this document guides you through the diagnostic steps to determine the specific root cause.

## Remedy

Follow these steps to detect the source of the problem:

### Step 1: Check the status of the Eventing backend CR

1. Check the Eventing backend CR. Is the field **EVENTINGREADY** `true`?
   
    ```bash
    kubectl -n kyma-system get eventingbackends.eventing.kyma-project.io
    ```

2. If **EVENTINGREADY** is `false`, check the exact reason of the error in the status of the CR by running the command:

    ```bash
    kubectl -n kyma-system get eventingbackends.eventing.kyma-project.io eventing-backend -o yaml
    ```

3. If **EVENTINGREADY** is `true`, the Eventing backend CR is not an issue. Follow the next steps to find the source of the problem.

### Step 2: Check the status of the Subscription

1. Check whether the Subscription is ready. Run the command:

    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME}
    ```

2. If the Subscription is not ready, check the exact reason of the error in the status of the Subscription by running the command:

    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o yaml
    ```

    If the status of the Subscription informs you that the sink is not a valid cluster local svc, investigate the Subscription sink. Verify if the sink is a valid HTTP endpoint, for example: `test.test.svc.cluster.local`.

3. Check if the event type defined in the Subscription is correctly formatted as specified in the [event names](../../05-technical-reference/evnt-01-event-names.md) guidelines.
   Also, check if the event type is using the correct `eventTypePrefix`. The event type must start with the `eventTypePrefix`. Run the following command to get the configured `eventTypePrefix` in Eventing Services:

    ```bash
    kubectl get configmaps -n kyma-system eventing -o jsonpath='{.data.eventTypePrefix}'
    ```

### Step 3: Check if the event was published correctly

1. Check the HTTP status code returned after sending an event.

    - If the HTTP status code is 4xx, check if you are sending the events in correct formats. Eventing supports two event formats (legacy and cloud events); see the [Eventing tutorials](../../03-tutorials/00-eventing) for more information.

    -  If the HTTP status code is 5xx, check the logs from the Eventing publisher proxy Pod for any errors. To fetch these logs, run this command:
   
        ```bash
        kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=eventing-publisher-proxy
        ```

    - If the HTTP status code is 2xx but it's still not working, verify if you are sending the event on same event type as defined in the Subscription.
  
        ```bash
        kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o jsonpath='{.spec.filter.filters}'
        ```

### Step 4: Check the Eventing Controller logs

1. Check the logs from the Eventing Controller Pod for any errors and to verify that the event is dispatched.
   To fetch these logs, run this command:

    ```bash
    kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=controller
    ```

2. Check for any error messages in the logs. If the event dispatch log `"message":"event dispatched"` is not present for NATS backend, the issue could be one of the following:

   - The subscriber (the sink) is not reachable or the subscriber cannot process the event. Check the logs of the subscriber instance.

   - The event was published in a wrong format.

   - Eventing Controller cannot connect to NATS Server.

### Step 5: Check if the Subscription sink is healthy

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

### Step 6: Check NATS JetStream status

1. Check the health of NATS Pods. Run the command:

    ```bash
    kubectl -n kyma-system get pods -l nats_cluster=eventing-nats
    ```

2. Check if the stream and consumers exist in NATS JetStream by following the [JetStream troubleshooting guide](evnt-02-jetstream-troubleshooting.md).


If you can't find a solution, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to our [Slack channel](http://slack.kyma-project.io/) to get direct support from the community.
