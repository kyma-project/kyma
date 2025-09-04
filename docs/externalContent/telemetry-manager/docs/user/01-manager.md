# Telemetry Manager

As the core element of the Telemetry module, Telemetry Manager manages the lifecycle of other Telemetry module components by watching user-created resources.

## Module Lifecycle

The Telemetry module includes Telemetry Manager, a Kubernetes [operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) that's described by a custom resource of type Telemetry. Telemetry Manager has the following tasks:

1. Watch the module configuration for changes and sync the module status to it.
2. Watch for the user-created Kubernetes resources LogPipeline, TracePipeline, and MetricPipeline. In these resources, you specify what data of a signal type to collect and where to ship it.
3. Manage the lifecycle of the self monitor and the user-configured agents and gateways.
   For example, only if you defined a LogPipeline resource, the Fluent Bit DaemonSet is deployed as log agent.

![Manager](assets/manager-resources.drawio.svg)

### Self Monitor

The Telemetry module contains a self monitor, based on [Prometheus](https://prometheus.io/), to collect and evaluate metrics from the managed gateways and agents. Telemetry Manager retrieves the current pipeline health from the self monitor and adjusts the status of the pipeline resources and the module status.
Additionally, you can monitor the health of your pipelines in an integrated backend like [SAP Cloud Logging](./integration/sap-cloud-logging/README.md#use-sap-cloud-logging-alerts): To set up alerts and reports in the backend, use the [pipeline health metrics](./04-metrics.md#5-monitor-pipeline-health) emitted by your MetricPipeline.

![Self-Monitor](assets/manager-arch.drawio.svg)

## Module Configuration and Status

For configuration options and the overall status of the module, see the specification of the related [Telemetry resource](./resources/01-telemetry.md).
