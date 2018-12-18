---
title: Subscription updates
type: Details
---

To update the subscription CRD, run this command:

`kubectl edit crd subscriptions.eventing.kyma-project.io`

The Event Bus reacts to the changes in subscription CRD, and updates the corresponding NATS-Streaming subscription accordingly.

>**Note**: The current subscription update mechanism recreates a subscription with new specifications. This may result in the loss of messages delivered during the recreation process.