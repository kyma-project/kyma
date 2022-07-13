---
title: Subscriber receives irrelevant events
---

## Symptom

Subscriber receives irrelevant events. 

## Cause

To conform to Cloud Event specifications, Eventing modifies the event names to filter out non-alphanumeric characters. For details, see [event name cleanup](../../05-technical-reference/evnt-01-event-names.md#event-name-cleanup).
In some cases, it can lead to a naming collision, which can cause subscribers to receive irrelevant events.

## Remedy

Follow these steps to detect if naming collision is the source of the problem:

1. Get the clean event types from the status of the Subscription.
 
    ```bash
    kubectl -n {NAMESPACE} get subscriptions.eventing.kyma-project.io {NAME} -o jsonpath='{.status.cleanEventTypes}'
    ```

2. Search for any other Subscription using the same `CleanEventType` as in your Subscription.
    
    ```bash
    kubectl get subscriptions.eventing.kyma-project.io -A | grep {CLEAN_EVENT_TYPE}
    ```
    
3. If you find that the `CleanEventType` collides with some other Subscription, a solution for this would be to provide an `application-type` label (with alphanumeric characters only), which is then used by the Eventing services instead of the Application name. If the `application-type` label also contains any non-alphanumeric character, the underlying Eventing services clean it and use the cleaned label. 
