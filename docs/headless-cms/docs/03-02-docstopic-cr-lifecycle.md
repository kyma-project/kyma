---
title: DocsTopic custom resource lifecycle
type: Details
---

>**NOTE:** This lifecycle also applies to the ClusterDocsTopic CR.

## Asset CR manual changes

The DocsTopic custom resource (CR) coordinates Asset CR creation, deletion, and modifications. The DocsTopic Controller verifies DocsTopic definition on a regular basis and creates, deletes, or modifies Assets CRs accordingly.

The DocsTopic CR acts as the only source of truth for the Asset CRs it orchestrates. If you modify or remove any of them manually, DocsTopic Controller automatically overwrites such an Asset CR or updates it based on the DocsTopic CR definition.

##  DocsTopic CR and Asset CR dependencies

Asset CRs and DocsTopic CRs are also interdependent in terms of names, definitions, and statuses.

### Names

The name of every Asset CR created by the DocsTopic Controller consists of these three elements:

- the name of the DocsTopic CR, such as `service-catalog`.
- the source type of the given asset in the DocsTopic CR, such as `asyncapi`.
- a randomly generated string, such as `1b38grj5vcu1l`.

The full name of such an Asset CR that follows the **{docsTopic-name}-{asset-source}-{suffix}** pattern is **service-catalog-asyncapi-1b38grj5vcu1l**.

### Labels

There are two labels in every Asset CR created from DocsTopic CRs. Both of them are based on DocsTopic CRs definitions:

- **cms.kyma-project.io/type** equals a given **type** parameter from the DocsTopic CR, such as `asyncapi`.

- **cms.kyma-project.io/docs-topic** equals the **name** metadata from the DocsTopic CR, such as `service-catalog`.

### Statuses

The status of the DocsTopic CR depends heavily on the status phase of all Asset CRs it creates. It is:

- `Ready` when all related Asset CRs already are in the `Ready` phase.
- `Pending` when it awaits the confirmation that all related Asset CRs are in the `Ready` status. If any Asset CR is in the `Failed` status, the status of the DocsTopic CR remains `Pending`.
- `Failed` when processing of the DocsTopic CR fails. For example, the DocsTopic CR can fail if you provide incorrect or duplicated data in its specification.
