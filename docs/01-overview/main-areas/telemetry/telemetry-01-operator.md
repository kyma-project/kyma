---
title: Telemetry - Operator
---

## Module lifecycle

The module on its own ships a single component only, namely the Telemetry Operator. The operator implements the Kubernetes controller pattern and manages the whole lifecycle of all other components relevant for this module. The operator watches for Kubernetes resources created by the user of type LogPipeline, TracePipeline and in future MetricPipeline. With these, the user describes in a declarative way what data of a signal type to collect and where to ship it.
If the operator detects a configuration, it will on demand roll-out the relevant collector components.

![Operator](./assets/operator-lifecycle.drawio.svg)

## Configuration

At the moment the operator has no configuration options. As part of the [modularization](https://github.com/kyma-project/kyma/issues/16301) efforts it is planned to introduce a dedicated module resource to watch the status of the module but also to provide advanced configuration options for the managed components.
