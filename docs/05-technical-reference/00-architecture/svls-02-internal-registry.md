---
title: Internal Docker Registry
---

Kyma serverless module by default comes with internal docker registry deployment that is used to store function container images w/o necessity to  use 3rd party registry

Internal docker  is not recommended for production, as

is not deployed in HA setup
has limited storage space and no garbage collection of orphaned images
Still, it is very convenient for development and getting first time experience of kyma serverless.

The following diagram illustrates how it works.

![Serverless architecture](./assets/svls-internal-registry.svg)

1. Build job pushes function image to docker registry using in-cluster URL
2. Internal docker registry URL is resolved by K8s DNS to real ip address
3. Kubelet fetch image using URL: `localhost:{node_port}/{image}`
4. NodePort allows kubelet to get into the cluster network and translate `localhost` to `internal-registry.kyma-system.svc.cluster.local` and ask the k8s dns to resolve the name
5. K8s DNS service resolve the name and provide the IP of `internal docker registry`

**NOTE:** Kubelet cannot resolve in-cluster URL that's why serverless uses NodePort service.
**NOTE:** NodePort service routing assure that pull request reaches the internal docker registry regardless of whether it is from different node.