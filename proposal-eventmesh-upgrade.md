# Proposal: migration to Knative Event Mesh

## Contents

1. [Expectation](#expectation)
2. [Steps](#steps)
   * [First Kyma update](#first-kyma-update)
     1. [Set up new event mesh for existing Applications](#1-set-up-new-event-mesh-for-existing-applications)
     2. [Deploy compatibility layer](#2-deploy-compatibility-layer)
     3. [Drain legacy message channels](#3-drain-legacy-message-channels)
     4. [Purge event-bus component](#4-purge-event-bus-component)
   * [Second Kyma update](#second-kyma-update)

## Expectation

At the end of the Kyma upgrade, the user is left with an event mesh that has feature parity with the previous Kyma
version, without the need for any manual configuration and without message loss.

## Steps

The upgrade happens in two steps to avoid creating a major disruption and to keep the logical path a little clearer. In
the first step, we ensure existing Kyma applications have a backing event mesh in place and delete the `event-bus`
component alone. In a second step, we remove leftovers from all other charts and rely on the natural deletion (via Helm)
that will happen upon the next Kyma update.

### First Kyma update

_target: 1.11_

#### 1. Set up new event mesh for existing Applications

**Logical**

Existing eventing components get replicated onto the new event mesh, for each Kyma application. As a side effect, each
application can technically support both event meshes, old and new, for the duration of this first step.

**Technical**

The Kyma operator proceeds component by component, iteratively. The most logical place to hook update logic is at the
level of the `event-bus` component, although the eventing in Kyma encompasses multiple components.

We are not expecting updates to the `event-bus` component to occur, so any Helm upgrade hook could be used to run this
job. For the rest of the procedure, we will agree on using the `pre-upgrade` hook.

At the end of this step, the update continues only if the `pre-upgrade` hook has succeeded.

#### 2. Deploy compatibility layer

**Logical**

Messages sent to legacy APIs get routed to the new event mesh.

**Technical**

_To be defined (may be part of an existing component, or a separate component)_

#### 3. Drain legacy message channels

**Logical**

In-flight messages sent to legacy APIs must be delivered before the old event mesh gets shut down.

**Technical**

_To be defined: watch delivery metrics / something else?_

#### 4. Purge event-bus component

**Logical**

The `event-bus` component gets uninstalled from the cluster, including the control plane (chart) and data plane (API
objects).

**Technical**

The Kyma operator does not currently uninstall components which are no longer required, so we need to implement custom
logic at the end of the operator's reconciliation to purge the `event-mesh` chart and remove its entry from the
Installation object.

The `event-bus` component does not include any CRD, so there is no risk to de-register a resource before all its objects
have been deleted.

### Second Kyma update

_target: 1.12_

Remove leftovers from other Kyma components:

* cluster-essentials:
  * Subscription CRD
  * EventActivation CRD

_Others?_
