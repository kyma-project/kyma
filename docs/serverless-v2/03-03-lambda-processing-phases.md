---
title: Lambda processing phases
type: Details
---

From the moment you create or update a lambda (Function CR) until the time it reaches the final `Running` status phase, lambda goes through three processing phases:

1. `Initializing`
2. `Building`
3. `Deploying`

The diagrams illustrate these three core phases of a lambda processing circle that the Function Controller handles. It also lists all custom resources involved in this process and specifies in which cases their update is required.

>**NOTE:** Before you start reading, see the [Function CR](#custom-resource-function) document for its detailed definition, description of all lambda status phases, its substages, and reasons.

## Initializing

This initial phase starts when you create a Function CR with configuration specifying the lambda's setup. It ends with creating a ConfigMap and a TaskRun CR ready for the lambda image built.

Updating an existing lambda involves an image rebuild only if any previous lambda processing phase failed (**phase: Failed**), or if you change lambda's configuration, so its body (**function**) or dependencies (**deps**). An update of lambda's environmental variables doesn't require image rebuild, and it only affects KService in the `Deploying` phase.

> **NOTE:** Each time you update lambda's configuration, the Function Controller deletes all previous TaskRun CRs for the given lambda's **UID** each time you update lambda's configuration.

![Initializing stage](./assets/initializing.svg)

## Building

This phase involves fetching and processing the TaskRun CR. It ends successfully when the lambda image is built and sent to the Docker registry. If the image already existed and only update is required, the Docker image receives a new tag.

![Building stage](./assets/building.svg)

## Deploying

This stage revolves around creating a KService or updating it when you previously changed environment variables in the Function CR or the image was rebuilt. In general, the KService is considered updated when both environment variables and the image tag in the KService are up to date. Thanks to the implemented reconciliation loop, the Function Controller constantly observes all newly created or updated KServices. If it detects one, it fetches its status and only then updates the lambda's status phase to `Running`.

![Deploying stage](./assets/deploying.svg)
