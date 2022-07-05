---
title: Subscriber receiving irrelevant Events
---

## Symptom(s)

Subscriber is receiving irrelevant Events. 

## Remedy

As described in [Event Names guildelines](../../../05-technical-reference/evnt-01-event-names.md), sometimes Eventing must modify the event names to filter out non-alphanumeric character to conform to Cloud Event specifications.
In some cases, it can lead to a naming collision which can result into Subscribers receiving irrelevant Events.

Follow these steps to detect if it is the source of the problem:

1. Get the Clean Event Types from the status of the Subscription.

    Run the following command to get the `CleanEventType` of the Subscription:
    
    ```bash
    kubectl -n {NAMESPACE} get subscription {NAME} -o jsonpath='{.status.cleanEventTypes}'
    ```

2. Search for any other Subscription using the same `CleanEventType` as in your Subscription.

    Run the command:
    
    ```bash
    kubectl get subscriptions -A | grep {CLEAN_EVENT_TYPE}
    ```
    
    If you find that the `CleanEventType` is colliding with some other Subscription then a solution could be use a different Application or Event type name for your Subscription. 
    Check out [Event Names guildelines](../../../05-technical-reference/evnt-01-event-names.md) for more information.
