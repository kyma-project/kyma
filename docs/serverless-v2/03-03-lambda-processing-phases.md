---
title: Lambda processing phases
type: Details
---

From the moment you create or update a lambda (Function CR) until the time it reaches the final `Running` status phase, the lambda goes through three processing phases:

1. `Initializing`
2. `Building`
3. `Deploying`

The diagrams illustrate these three core phases of a lambda processing circle that the Function Controller handles. They also list all custom resources involved in this process and specify in which cases their update is required.

>**NOTE:** Before you start reading, see the [Function CR](#custom-resource-function) document for the custom resource detailed definition, the list of all lambda status phases and reasons for their success or failure.

## Initializing

This initial phase starts when you create a Function CR with configuration specifying the lambda's setup. It ends with creating a ConfigMap and a TaskRun CR ready for building the lambda image.

Updating an existing lambda requires an image rebuild only if any previous lambda processing phase failed (**phase: Failed**), or if you change lambda's body (**function**) or dependencies (**deps**). An update of lambda's other configuration details, such as environment variables, replicas, resources, or labels, doesn't require image rebuild, and it only affects KService in the `Deploying` phase.

> **NOTE:** Each time you update lambda's configuration, the Function Controller deletes all previous TaskRun CRs for the given lambda's **UID**.

![Initializing stage](./assets/initializing.svg)

## Building

This phase involves fetching and processing the TaskRun CR. It ends successfully when the lambda image is built and sent to the Docker registry. If the image already existed and only update was required, the Docker image receives a new tag.

![Building stage](./assets/building.svg)

## Deploying

This stage revolves around creating a KService or updating it when configuration changes were made in the Function CR or the lambda image was rebuilt. In general, the KService is considered updated when both configuration and the image tag in the KService are up to date. Thanks to the implemented reconciliation loop, the Function Controller constantly observes all newly created or updated KServices. If it detects one, it fetches the KService status and only then updates the lambda's status phase to `Running`.

![Deploying stage](./assets/deploying.svg)
