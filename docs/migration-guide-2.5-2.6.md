---
title: Migration Guide 2.5-2.6
---

Due to the enablement of the telemetry component by default, some logging resources must be deleted. When you upgrade from Kyma 2.5 to 2.6, either run the script [2.5-2.6-cleanup-logging-fluent-bit.sh](./assets/2.5-2.6-cleanup-logging-fluent-bit.sh) or run the commands from the script manually.

If you are using override values for the logging component to configure additional Fluent Bit filters, parsers, or outputs, you can apply them by creating a LogPipeline or LogParser custom resource. For more information, see [Kyma's telemetry component](./01-overview/main-areas/observability/obsv-04-telemetry-in-kyma.md). Other overrides to Fluent Bit must be applied to the telemetry component now.

Kyma will continue to maintain the Fluent Bit setup from the logging component until release 2.8. If you want to keep your existing override values for now, you can re-enable Fluent Bit in the logging component by deploying Kyma with the following arguments:

```bash
kyma deploy --value=global.telemetry.enabled=false --value=logging.fluent-bit.enabled=true
```