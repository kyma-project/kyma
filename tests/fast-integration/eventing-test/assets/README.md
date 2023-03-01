# Eventing Sink

## Overview

Use the `eventing-sink` Function to test the end-to-end flow for Eventing. 
It performs three tasks:
1. Forward requests to publish events to event-publisher-proxy.
2. Save all received events in memory.
3. Return saved events to users.

## Usage

### Forward requests to publish events to `event-publisher-proxy`

To publish an event to `event-publisher-proxy`, send an HTTP request with parameter `send=true` and the following body:

```
{
    url: `http://eventing-event-publisher-proxy.kyma-system/publish`,
    data: {
        headers: {}, // Event headers
        payload: {}  // Event payload
    },
}
```

### Fetch an event received by `eventing-sink` from a backend

To fetch an event received by `eventing-sink`, send an HTTP request with parameter `eventid=<id>`. The response contains:

```
{
    success: true|false,               // Will be false if requested event was not found.
    event: {},                         // Event object.
    metadata: {
        podName: "eventing-sink-xxx"   // Eventing sink pod name.
    }
}
```
