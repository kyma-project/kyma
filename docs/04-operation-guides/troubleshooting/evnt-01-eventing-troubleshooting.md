---
title: Did not receive an event - basic diagnostics
---

## Symptom

You publish an event but the event is not received by the subscriber.

## Remedy

Follow these steps to detect the source of the problem:

1. Check the eventing backend CRD, if the field `EVENTINGREADY` is true:
   
   ```bash
   kubectl -n kyma-system get eventingBackend
   ```
   If `EVENTINGREADY` is false, check the exact reason of the error in the status of the CRD by running the command:

   ```bash
   kubectl -n kyma-system get eventingBackend eventing-backend -o yaml
   ```

2. Check the subscription, if it is ready. Run the command:

   ```bash
   kubectl get subscription -A
   ```
   If the subscription is not ready, check the exact reason of the error in the status of the subscription by running the command:

   ```bash
   kubectl -n {NAMESPACE} get subscription {NAME} -o yaml
   ```

3. Check the controller logs.

   Check the logs from the Eventing Controller Pod for any errors or to verify that the event is dispatched.
   To fetch these logs, run this command:

   ```bash
   kubectl -n kyma-system logs -l app.kubernetes.io/instance=eventing,app.kubernetes.io/name=controller
   ```
   
   If the event dispatch log `"message":"event dispatched"` is not present for NATS backend, then the issue could lie with how the event was published.