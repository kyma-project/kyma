---
title: Basic troubleshooting
type: Troubleshooting
---

## Lambda does not receive Events

If the lambda or the service does not receive any Events, do the following:

  - Confirm that the EventActivation custom resource is in place.
  - Ensure that the webhook defined for the lambda or the service is up and
    running.
  - Make sure the Events are published.

## Errors while sending Events

If errors appear while sending Events, do the following:

  - Check if the `publish` application is up and running.
  - Make sure that NATS Streaming is up and running.

