---
title: Kyma Eventing - Basic Diagnostics
---

## Symptom(s)

- You publish an event but the event is not received by the subscriber.

- Subscription is not in ready state.

- You are unable to publish events.

## Remedy

Follow these steps to detect the source of the problem:

### Step 1: Check the status of Check the Eventing backend CR

Check the Eventing backend CR, if the field `EVENTINGREADY` is true:
   
```bash
kubectl -n kyma-system get eventingBackend
```
If `EVENTINGREADY` is false, check the exact reason of the error in the status of the CR by running the command:

```bash
kubectl -n kyma-system get eventingBackend eventing-backend -o yaml
```

### Step 2: Check the status of the Subscription

Check the Subscription, if it is ready. Run the command:

```bash
kubectl -n {NAMESPACE} get subscription {NAME}
```
If the Subscription is not ready, check the exact reason of the error in the status of the Subscription by running the command:

```bash
kubectl -n {NAMESPACE} get subscription {NAME} -o yaml
```

If there is a problem with the Subscription sink you will see the message that sink is not valid cluster local svc. 
Verify if the sink is a valid HTTP endpoint, for example: `test.test.svc.cluster.local`.

Check if the Event type defined the Subscription is correctly formatted as specified in the [guidelines](../../05-technical-reference/evnt-01-event-names.md).
Also, check if the Event type is using correct `eventTypePrefix`, the event type has to start with the `eventTypePrefix`, otherwise the Subscription won't be ready.

Run the following command to get the configured `eventTypePrefix` in Eventing Services:
```bash
kubectl get configmaps -n kyma-system eventing -o jsonpath='{.data.eventTypePrefix}'
```

### Step 3: Check if the Subscription sink is healthy.

Check the workload Url defined in the Subscription sink is correct and healthy to receive events.
To get the sink from the Subscription, run this command:

```bash
kubectl -n {NAMESPACE} get subscription {NAME} -o jsonpath='{.spec.sink}'
```

You can run the following commands to check the health of the sink:

```bash
kubectl -n default run --image=curlimages/curl --restart=Never sink-test-tmp -- curl --head {SINK_URL}
kubectl -n default logs sink-test-tmp 
kubectl -n default delete pod sink-test-tmp
```

### Step 4: Check the eventing-controller logs.

Check the logs from the Eventing Controller Pod for any errors or to verify that the event is dispatched.
To fetch these logs, run this command:

```bash
kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=controller
```

Check for any error messages in the logs. 

If the event dispatch log `"message":"event dispatched"` is not present for NATS backend, then the issue could be:

1. The subscriber (i.e. sink) is not reachable.

2. The event was not published in correct format.

3. Eventing-controller is not able to connect to NATS Server.

### Step 5: Check if the Event was published correctly

Check if sending an event returns an HTTP Status Code other than 2xx. 

- If the HTTP Status Code was 4xx, then check if you are sending the Events in correct formats. Eventing supports two Event formats (i.e. legacy and cloud events), see the Eventing [tutorials](../../03-tutorials/00-eventing) section for more information.

- If the HTTP Status Code was 5xx, then check the logs from the Eventing Publisher Proxy Pod for any errors. To fetch these logs, run this command:
    ```bash
    kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=eventing-publisher-proxy
    ```
  
- If the HTTP Status Code was 2xx, then verify if you are sending the event on same Event type as defined in the subscription.
  ```bash
  kubectl -n {NAMESPACE} get subscription {NAME} -o jsonpath='{.spec.filter.filters}' | jq
  ```

### Step 6: Check NATS JetStream status

Check the health of NATS Pods. Run the command:

```bash
kubectl -n kyma-system get pods -l nats_cluster=eventing-nats
```

Check if the stream and consumers exists in NATS JetStream by following this [JetStream troubleshoot guide](./evnt-02-jetstream-troubleshooting.md).







