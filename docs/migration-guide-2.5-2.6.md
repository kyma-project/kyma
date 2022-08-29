---
title: Migration Guide 2.5-2.6
---

# Migrate Fluent Bit

With this release, the Telemetry component is enabled by default; so Fluent Bit is now managed by the Telemetry component instead of the Logging component. Hereby, the configuration of Fluent Bit is done at runtime using the new LogPipeline CRD, instead of overrides at installation time.

The migration steps depend on whether you are using override values for the Logging component to configure additional Fluent Bit filters, parsers, or outputs.

If you use **no override values** for Fluent Bit configuration, skip steps 1-2 and go directly to step 3.

If you are using override values to configure Fluent Bit, we recommend that you avoid downtime by temporarily re-enabling Fluent Bit in the Logging component while you set up the new configuration. You must finish the migration until Kyma release 2.8, afterwards Kyma will stop maintaining the legacy setup.

1. To keep using your existing setup temporarily, re-enable Fluent Bit in the Logging component by deploying Kyma with the following arguments:
   
   ```bash
   kyma deploy --value=global.telemetry.enabled=false --value=logging.fluent-bit.enabled=true
   ```

2. To apply your override values in the Telemetry component, create a LogPipeline or LogParser custom resource. For more information, see [Kyma's telemetry component](./01-overview/main-areas/observability/obsv-04-telemetry-in-kyma.md). Other overrides to Fluent Bit must be applied to the Telemetry component now.

   After setting up your configuration in the Telemetry component, you don't need the old override values anymore, so you can delete the obsolete Fluent Bit resources.

3. Delete the affected logging resources: Either run the script [2.5-2.6-cleanup-logging-fluent-bit.sh](./assets/2.5-2.6-cleanup-logging-fluent-bit.sh) or run the commands from the script manually.
