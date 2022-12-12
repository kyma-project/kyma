---
title: What is Telemetry in Kyma?
---

Fundamentally, ["Observability"](https://opentelemetry.io/docs/concepts/observability-primer/) is a measure of how well the internal states of single components can be reflected by the application's external outputs. The insights that an application and the surrounding infrastructure exposes are displayed in the form of metrics, traces, and logs - collectively, that's called "telemetry" or ["signals"](https://opentelemetry.io/docs/concepts/signals/). These can be exposed by employing modern instrumentation.

In order to implement Day-2 operations for a distributed application running in a container runtime, the single components of an application needs to expose these signals by employing modern instrumentation. Furthermore the signals needs to get collected and enriched with the infrastructural metadata in order to ship them to a target system.

![Stages of Observability](./assets/stages.drawio.svg)

There a plenty of observability backends available either as service or as a self-manageable solution focussing on different aspects and scenarios. Here, one solution will never fit all sizes and the need to integrate with a specific solution will always be present. That's why the aspect of instrumenting and shipping your telemetry instance in an easy way in a vendor-neutral way is relevant for Kyma in order to enable observability for your application with low effort by integration into existing backends. That aspects must happen alongside with your application and here managed tooling in combination with guidance can provide the biggest effect for users on initial investment and maintenance effort. Also, Kyma will not focus on providing a managed in-cluster backend solution as solution for an enterprise-grade setup will demand a central outer-cluster solution.

The Telemetry module focuses exactly on these aspects (instrumentation/collection/shipment) happening in the runtime and explicitly de-focus backends. Tutorials are provided additional to install lightweight in-cluster backends for demo or development purposes.

## Features

The Telemetry module is enabling your application with telemetry support by providing:

- guidance for the instrumentation - based on the [Open Telemetry](https://opentelemetry.io/) community samples are provided on how to instrument your code using the [Open Telemetry SDKs](https://opentelemetry.io/docs/instrumentation/) in nearly any programming language
- tooling for collection, filtering and shipment - based on the [Open Telemetry Collector](https://opentelemetry.io/docs/collector/), you can configure basic pipelines to filter and ship telemetry data
- integration in a vendor-neutral way to a vendor specific observability system - based on the [OpenTelemetry protocol (OTLP)](https://opentelemetry.io/docs/reference/specification/protocol/) backend systems can be integrated
- opt-out from features for advanced scenarios - at anytime you can opt-out per data type and realize the telemetry data collection/shipment with custom tooling
- SAP BTP as first class integration - integration into BTP Observability services will be prioritized
- Enterprise-grade qualities - the setup is battle-tested and will satisfy typical development standards

Initially,
- metrics will not be supported, follow the related [epic](https://github.com/kyma-project/kyma/issues/13079) for tracking the progress of the minimal first version.
- logs will be not based on the vendor-neutral OTLP protocol, follow this [epic](https://github.com/kyma-project/kyma/issues/16307) to understand the current progress on that
- the focus on filtering capabilities is to reduce the overall shipment bandwidth to an acceptable minimum, as telemetry data is high-volume data which is always related to a price. So filtering features are focused on reducing signal data to relevant namespaces or workloads for example. By default, system related telemetry data will not be shipped. For custom scenarios there is always the opt-out possibility.

## Scope

There will be a focus on the signals of application logs, distributed traces and metrics only. Other kind of signals are not considered. Also, logs like audit logs or operational logs are specifically de-scoped.

Supported integration scenarios are neutral to the vendor of the target system, a vendor-specific way is not planned at the moment.

## Components

![Components](./assets/components.drawio.svg)

### Telemetry Operator

The module ships the Telemetry Operator as it's heart component. The operator implements the Kubernetes controller pattern and manages the whole lifecycle of all other components covered in the module. The operator watches for resources created by the user of type LogPipeline, TracePipeline and in future MetricPipeline. With these, the user describes in a declarative way what data of a signal type to collect and where to ship it.
If the operator detects a configuration, it will on demand roll-out the relevant collector components.
More details to the operator itself can be found at the [operator](./telemetry-01-operator.md) page.

### Log Collector

The Log Collector is based on a [Fluentbit](https://fluentbit.io/) installation running as a [DamonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/). It reads all logs of the containers in the runtime and ships them according to a LogPipeline configuration. More details can be found at the detailed section about [Logs](./telemetry-02-logs.md).

### Trace Collector

The Trace Collector is based on a [Otel Collector](https://opentelemetry.io/docs/collector/) [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/). It provides an [OTLP](https://opentelemetry.io/docs/reference/specification/protocol/) based endpoint where applications can push trace signals to. According to a TracePipeline configuration the collector will process and ship the trace data to a target system. More details can be found at the detailed section about [Traces](./telemetry-03-traces.md)

### Metrics Collector

To be implemented [soon](https://github.com/kyma-project/kyma/issues/13079).
