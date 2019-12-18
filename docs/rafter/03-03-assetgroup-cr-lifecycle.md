---
title: AssetGroup custom resource lifecycle
type: Details
---

>**NOTE:** This lifecycle also applies to the ClusterAssetGroup CR.

## Asset CR manual changes

The AssetGroup custom resource (CR) coordinates Asset CR creation, deletion, and modification. The AssetGroup Controller (AGC) verifies AssetGroup definition on a regular basis and creates, deletes, or modifies Asset CRs accordingly.

The AssetGroup CR acts as the single source of truth for the Asset CRs it orchestrates. If you modify or remove any of them manually, the AGC automatically overwrites such an Asset CR or updates it based on the AssetGroup CR definition.

##  AssetGroup CR and Asset CR dependencies

Asset CRs and AssetGroup CRs are also interdependent in terms of names, definitions, and statuses.

### Names

The name of every Asset CR created by the AGC consists of these three elements:

- The name of the AssetGroup CR, such as `service-catalog`.
- The source type of the given asset in the AssetGroup CR, such as `asyncapi`.
- A randomly generated string, such as `1b38grj5vcu1l`.

The full name of such an Asset CR that follows the `{assetGroup-name}-{asset-source}-{suffix}` pattern is `service-catalog-asyncapi-1b38grj5vcu1l`.

### Labels

There are two labels in every Asset CR created from AssetGroup CRs. Both of them are based on AssetGroup CRs definitions:

- **rafter.kyma-project.io/type** equals a given **type** parameter from the AssetGroup CR, such as `asyncapi`.

- **rafter.kyma-project.io/asset-group** equals the **name** metadata from the AssetGroup CR, such as `service-catalog`.

### Statuses

The status of the AssetGroup CR depends heavily on the status phase of all Asset CRs it creates. It is:

- `Ready` when all related Asset CRs are already in the `Ready` phase.
- `Pending` when it awaits the confirmation that all related Asset CRs are in the `Ready` phase. If any Asset CR is in the `Failed` phase, the status of the AssetGroup CR remains `Pending`.
- `Failed` when processing of the AssetGroup CR fails. For example, the AssetGroup CR can fail if you provide incorrect or duplicated data in its specification.
