# Proposal: migration to Knative Event Mesh

## Contents

1. [Expectation](#expectation)
2. [Logical steps](#logical-steps)
3. [Technical considerations](#technical-considerations)

## Expectation

At the end of the Kyma upgrade, the user is left with an event mesh that has feature parity with the previous Kyma
version, without the need for any manual configuration and without message loss.

## Logical steps

1. **Replicate existing eventing components onto the new event mesh, for each Kyma application.**
    For the duration of the update, each application can technically support both event meshes, old and new.
2. **Deploy compatibility layer** (_to be defined_)
    Messages sent to legacy APIs get routed to the new event mesh.
3. **Drain legacy message channels**
    In-flight messages sent to legacy APIs must be delivered before the old event mesh gets shut down.
4. **Purge old event mesh**
    Includes both control (chart) and data (API objects) planes.

## Technical considerations

* **The update logic for _step 1_ needs to be hooked into an existing component, but the event mesh encompasses multiple components.**
    Proposed action: consolidate all custom update steps into update hooks at the level of the `event-bus` component,
    which ensures the component gets removed only if all previous operations succeed.
* **The Kyma operator does not currently uninstall components which are no longer required.**
    Proposed action: implement custom logic at the end of the operator's reconciliation to purge the `event-mesh` chart
    and remove its entry from the Installation object, since this will not be supported anymore moving forward.
* **Old API objects need to be removed before their respective CRDs are de-registered from the cluster.**
    This must happen before uninstalling the chart.
