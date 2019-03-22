---
title: DocsTopic custom resource lifecycle
type: Details
---

## Asset CR manual changes

The DocsTopic custom resource (CR) coordinates Asset CR creation, deletion, and changes. The DocsTopic Controller verifies DocsTopic definition on a regular basis and adds, removes, or modifies the Assets CRs accordingly.

The DocsTopic CR acts as the only source of truth for the Asset CRs it orchestrates. If you modify or remove any of them manually, DocsTopic Controller automatically overwrites it or recreates it based on the DocsTopic CR definition.

##  DocsTopic CR and Asset CR dependencies

Asset CRs and DocsTopic CRs are also interdependent in terms of names, definitions, and statuses.

### Names

The name of every Asset CR created by the DocsTopic Controller consists of these three elements:

- the name of the DocsTopic CR, such as `service-catalog`.
- the source of the given asset in the DocsTopic CR, such as `asyncapi`.
- randomly generated string, such as `1b38grj5vcu1l`.

The full name of such an Asset CR following the **{docsTopic-name}-{asset-source}-{suffix}** pattern is **service-catalog-asyncapi-1b38grj5vcu1l**.

### Labels

There are two labels in every Asset CR created from DocsTopic CRs that are based on these DocsTopic CRs definitions:

- the **type** label that equals a given **sources** parameter from the DocsTopic CR. For example, that is `markdown`.

- the **docsTopic** label equals the **name** metadata from the DocsTopic CR. For example, that is `service-catalog`.

## Statuses

The status of the DocsTopic CR heavily depends on the status phase of all Asset CRs it creates. It is:

- `Ready` when all related Asset CRs already are in the `Ready` phase.
- `Pending` when it awaits the confirmation of the statuses of all related Asset CRs.
- `Failed` when the processing of the DocsTopic CR fails.
