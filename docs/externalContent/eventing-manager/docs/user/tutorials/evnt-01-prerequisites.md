# Eventing Tutorials

Browse the tutorials for Eventing to learn how to use it step-by-step in different scenarios.

## Prerequisites

To perform the Eventing tutorials, you need the following setup:

1. [Kyma cluster provisioned](https://kyma-project.io/#/02-get-started/01-quick-install).

2. (Optional) [Kyma dashboard](https://kyma-project.io/#/01-overview/ui/README?id=kyma-dashboard) deployed in the Kyma cluster. To access the Dashboard, run:

   ```bash
   kyma dashboard
   ```

   Alternatively, you can just use the `kubectl` CLI instead.

3. (Optional) [CloudEvents Conformance Tool](https://github.com/cloudevents/conformance) for publishing events.

   ```bash
   go install github.com/cloudevents/conformance/cmd/cloudevents@latest
   ```

   Alternatively, you can just use `curl` to publish events.
