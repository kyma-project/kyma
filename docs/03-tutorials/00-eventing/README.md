---
title: Eventing - tutorials
---

Browse the tutorials for Eventing to learn how to use it step-by-step in different scenarios.

## Prerequisites

To perform the steps described in the Eventing tutorials, you need the following: 

1. [Kyma cluster provisioned](../../04-operation-guides/operations/02-install-kyma.md).
2. (Optional) [Kyma Dashboard](../../01-overview/ui/ui-01-gui.md) deployed on the Kyma cluster. To access the Dashboard, run: 
   ```bash
   kyma dashboard
   ```
   Alternatively, you can just use the `kubectl` CLI instead.
3. (Optional) [CloudEvents Conformance Tool](https://github.com/cloudevents/conformance) for publishing events. 
   ```bash
   go install github.com/cloudevents/conformance/cmd/cloudevents@latest
   ```
   Alternatively, you can just use `curl` to publish events.