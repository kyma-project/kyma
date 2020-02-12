# Tekton Pipelines

## Overview

The Tekton Pipelines project provides K8s resources for declaring CI/CD pipelines.

Tekton Pipelines are **cloud-native**:

- They run on Kubernetes.
- They have Kubernetes clusters as a first-class type.
- They use containers as their building blocks.

Tekton Pipelines are **decoupled**:

- One [Pipeline](https://github.com/tektoncd/pipeline/blob/master/docs/pipelines.md) resource can be used to deploy to any K8s cluster.
- The [Tasks](https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md) which make up a Pipeline can easily be run in isolation.
- Resources such as Git repositories can easily be swapped between runs.

Tekton Pipelines are **typed**:

- The concept of typed resources means that you can easily swap out implementations for such resources as Images. For example, you can use [kaniko](https://github.com/GoogleContainerTools/kaniko) or [buildkit](https://github.com/moby/buildkit) to build images.

To learn more about the Tekton Pipelines, go to the [Tekton Pipelines repository](https://github.com/tektoncd/pipeline).
