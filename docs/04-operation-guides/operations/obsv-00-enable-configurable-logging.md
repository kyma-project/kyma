---
title: Enable the telemetry component (beta feature)
---

## Context

<!-- Briefly provide background information for the task so that the users understand the purpose of the task and what they will gain by completing the task correctly. This section should be brief and does not replace or recreate a concept topic on the same subject, although the context section might include some conceptual information.
-->

With Kyma's telemetry component, you can provide your custom configuration for Fluent Bit, for example, to integrate with vendors like [VMWare](https://medium.com/@shrishs/log-forwarding-from-fluent-bit-to-vrealizeloginsightcloud-9eeb14b40276) using generic outputs, or with any vendor via a [fluentd integration](https://medium.com/hepsiburadatech/fluent-logging-architecture-fluent-bit-fluentd-elasticsearch-ca4a898e28aa) using the forward output. This enables you to integrate Kyma's logging backend with your own observability systems outside the Kyma cluster.

The new telemetry component ships a new Fluent Bit installation alongside with the new telemetry component, which will support the dynamic configuration of the Fluent Bit. By default, it will not ship any logs.
The new `global.telemetry.enabled` flag on the logging component disables the existing Fluent Bit installation. Furthermore, it activates a new Log Pipeline resource to enable log shipment to the existing Loki stack.

To compare Kyma's classic logging setup with the extended functionality, see the [logging architecture document](../../05-technical-reference/00-architecture/obsv-02-architecture-logging.md).

> **CAUTION:** To enable the telemetry component, you must change your Kyma installation; you can't do it on the fly. This is a beta feature, proceed with caution.

## Prerequisites

<!-- Describes information that the user needs to know or things they need to do or have before starting the immediate task.
If it's more than one prerequisite, use an unordered list.
For example, specify the authorizations the user must have and what software (and versions) must be installed already.
 -->

- You have installed Kyma with the classic logging chart.

## Steps

<!-- Provide a series of steps needed to perform the task.
Use a numbered list with one number for each action that the users must take.
It's good practice to describe the result of the procedure so that the users can see they accomplished the task successfully.
Sometimes it's also very helpful to describe the result of a specific step (don't use a number for step results, just a new line below the step). Remember about appropriate indentation for this line.
If the task at hand is typically followed by another one, you can add a link to that other document as "Next Steps".
-->

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

<!-- Not mandatory, but recommended. Help the reader to be sure they accomplished the task successfully. -->

You have installed the telemetry component and replaced the classic Fluent Bit installation with one that contains the new telemetry component.

## Next Steps

To forward logs to your preferred vendor, provide a custom configuration.
