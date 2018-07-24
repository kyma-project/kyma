---
title: Troubleshooting
type: Details
---

* If the lambda or the service does not receive any Events, check the following:
  - Confirm that the EventActivation is in place.
  - Ensure that the webhook defined for the lambda or the service is up and running.
  - Make sure the Events are published.

* If errors appear while sending Events:
  - Check if the `publish` application is up and running.
  - Make sure that NATS Streaming is up and running.
