# Tekton Pipelines

## Overview

The Tekton Pipelines project provides k8s-style resources for declaring CI/CD-style pipelines.

Tekton Pipelines are **Cloud Native**:

- Run on Kubernetes
- Have Kubernetes clusters as a first class type
- Use containers as their building blocks

Tekton Pipelines are **Decoupled**:

- One Pipeline can be used to deploy to any k8s cluster
- The Tasks which make up a Pipeline can easily be run in isolation
- Resources such as git repos can easily be swapped between runs

Tekton Pipelines are **Typed**:

- The concept of typed resources means that for a resource such as an `Image`, implementations can easily be swapped out (e.g. building with [kaniko](https://github.com/GoogleContainerTools/kaniko) v.s. [buildkit](https://github.com/moby/buildkit))

To learn more about the Tekton Pipelines, go to the [Tekton Pipelines repository](https://github.com/tektoncd/pipeline).
