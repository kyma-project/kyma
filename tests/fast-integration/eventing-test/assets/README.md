# Eventing Sink

## Overview

This eventing sink function is used to test the end-to-end flow for Eventing. It is responsible for three tasks:
1. Forward requests to publish events to event-publisher-proxy.
2. Save all received events in memory.
3. Return saved events to users.

## Usage

### Forward requests to publish events to event-publisher-proxy.

To publish an event to event-publisher-proxy, send a Http request with params `send=true` and body:

```
{
    url: `http://eventing-event-publisher-proxy.kyma-system/publish`,
    data: {
        headers: {}, // Event headers
        payload: {}  // Event payload
    },
}
```

### Fetch an event received by eventing-sink from a backend.

To fetch an event received by eventing-sink, send a Http request with params `eventid=<id>`. The response contains:

```
{
    success: true|false,               // Will be false if requested event was not found.
    event: {},                         // Event object.
    metadata: {
        podName: "eventing-sink-xxx"   // Eventing sink pod name.
    }
}
```
