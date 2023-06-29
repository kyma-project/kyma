---
title: Kyma Dashboard
---

## Purpose

Kyma uses [Busola](https://github.com/kyma-project/busola) as a central administration dashboard, which provides a graphical overview of your cluster and deployments.

You can deploy microservices, create Functions, and manage their configurations. You can also use it to register cloud providers for additional services, create instances of these services, and use them in your microservices or Functions.

## Integration

Busola is a web-based UI for managing resources within Kyma or any Kubernetes cluster. It consists of separate micro front-end applications managed by the [Luigi framework](https://luigi-project.io/). Busola has a dedicated Node.js back end, which is a proxy for a [Kubernetes API server](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver).
