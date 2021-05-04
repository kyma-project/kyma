---
title: Search for traces
type: Tutorials
---

You can search traces using tags. Tags are key-value pairs configured for each service.
The full list of tags for a service from the details of that service's span.

For example, use these tags for `event-publish-service`:

* **event-type**
* **event-type-ver**
* **event-id**
* **source-id**

To search the traces, you can use either a single tag, such as `event-type="order.created"`, or multiple tags, such as `event-type="order.created" event-type-ver="v1"`.