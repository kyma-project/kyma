---
title: Telemetry Manager
---

## Module lifecycle

Kyma's Telemetry module ships Telemetry Manager as its core component. Telemetry Manager is a Kubernetes [operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) that is described by a custom resource of type Telemetry. Telemetry Manager implements the Kubernetes controller pattern and manages the whole lifecycle of all other components covered in the Telemetry module.
Telemetry Manager watches for the user-created Kubernetes resources: LogPipeline, TracePipeline, and, in the future, MetricPipeline. In these resources, you specify what data of a signal type to collect and where to ship it.
If Telemetry Manager detects a configuration, it rolls out the relevant components on demand.

![Manager](./assets/manager-lifecycle.drawio.svg)

## Configuration

At the moment, you cannot configure Telemetry Manager. As part of the [modularization](https://github.com/kyma-project/kyma/issues/16301) efforts, is planned to support configuration in the specification of the related Telemetry resource.
