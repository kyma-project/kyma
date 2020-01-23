# Proposal: migration to Knative Event Mesh

## Contents

1. [Expectation](#expectation)
2. [Steps](#steps)
   1. [Set up new event mesh for existing Applications](#1-set-up-new-event-mesh-for-existing-applications)
   2. [Deploy compatibility layer](#2-deploy-compatibility-layer)
   3. [Drain legacy message channels](#3-drain-legacy-message-channels)
   4. [Purge old event mesh](#4-purge-old-event-mesh)
3. [Clean up](#clean-up)

## Expectation

At the end of the Kyma upgrade, the user is left with an event mesh that has feature parity with the previous Kyma
version, without the need for any manual configuration and without message loss.

## Steps

The upgrade is performed in a single Kyma release. There is technically no deprecation phase for the old event mesh
solution, which will be removed entirely from the target cluster in case the upgrade procedure described in this
document runs successfully.

All the steps described below are orchestrated by the Kyma operator. During a Kyma upgrade, the operator proceeds
component by component, iteratively. The most rational place to hook our migration logic is at the level of the
`cluster-essentials` component, which is the [very first chart](./installation/resources/installer-cr.yaml.tpl#L13-L15)
defined in the `Installation` object, so we can clear the path from there for the actual components' upgrades.

Each of the steps described below is implemented as a Kubernetes `Job` triggered by a `pre-upgrade` Helm hook. A
predictable ordering is enforced using the `helm.sh/hook-weight` annotation.

A failed step restores the state of the cluster as it was when the step started.

### 1. Set up new event mesh for existing Applications

**Logical**

Existing eventing components get replicated onto the new event mesh, for each Kyma application. As a side effect, each
application can technically support both event meshes, old and new, for the duration of this first step.

**Technical**

Self-contained migration logic delivered as a container image that creates `Trigger` objects. No API object gets deleted
in this step.

### 2. Deploy compatibility layer

**Logical**

Messages sent to legacy APIs get routed to the new event mesh.

**Technical**

_To be defined (may be part of an existing component, or a separate component)_

### 3. Drain legacy message channels

**Logical**

In-flight messages sent to legacy APIs must be delivered before the old event mesh gets shut down.

**Technical**

To avoid over-engineering, we decided to simply observe a grace period of 60sec to allow the delivery of in-flight
events to complete before proceeding with the next step. Metrics exposed by data Channels will not be taken into
account, unprocessed events will be lost.

### 4. Purge old event mesh

**Logical**

Objects belonging to the old event mesh are removed from the cluster. All Applications use the new event mesh
exclusively.

**Technical**

Self-contained logic as a container that deletes all instances of the following resources from the cluster:

  * `Subscription` (Kyma)
  * `EventActivation`

The corresponding CRDs are then immediately removed.

## Clean up

Leftovers from the old event mesh are removed from all other charts by relying on the natural deletion (via Helm) that
will happen in the rest of the Kyma upgrade. Ideally, changes to all involved components should be merged in a single
commit to avoid leaving unused services behind at any point.
