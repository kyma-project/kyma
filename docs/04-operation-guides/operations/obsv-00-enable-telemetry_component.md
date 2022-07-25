---
title: Enable the telemetry component (alpha)
---

## Context

With Kyma's alpha [telemetry component](./../../01-overview/main-areas/observability/obsv-04-telemetry-in-kyma.md), you can provide your custom configuration for Fluent Bit. This enables you to integrate Kyma's logging backend with your own observability systems outside the Kyma cluster.

The telemetry component includes a new Fluent Bit installation, which supports the dynamic configuration of Fluent Bit. By default, it does not ship any logs.
The new `global.telemetry.enabled` flag on the logging component disables the existing Fluent Bit installation. Furthermore, it activates a new LogPipeline resource to enable log shipment to the existing Loki stack.

To compare Kyma's classic logging setup with the extended functionality, see the [logging architecture document](../../05-technical-reference/00-architecture/obsv-02-architecture-logging.md).

> **CAUTION:** To enable the telemetry component, you must change your Kyma installation; you can't do it on the fly. This is an alpha feature, proceed with caution.

## Prerequisites

- You have installed Kyma with the classic logging chart.

## Steps

1. Update your Kyma installation by adding the telemetry component and enabling global telemetry for the logging component. Run:

   ```bash
   kyma deploy --component=telemetry --component logging --value global.telemetry.enabled=true
   ```

2. Delete the old Fluent Bit installation by deleting the respective resources. Run:

   ```bash
   kubectl delete daemonset -n kyma-system logging-fluent-bit
   kubectl delete configmap -n kyma-system logging-fluent-bit
   kubectl delete servicemonitor -n kyma-system logging-fluent-bit
   ```

## Result

You have installed the telemetry component and replaced the classic Fluent Bit installation with one that contains the new telemetry component.

## Next Steps

To forward logs to your preferred vendor, provide a custom configuration.
