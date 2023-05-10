---
title: Internal Docker Registry
---

Internal docker registry is used for serverless to push built function images which can be used to deploy function without using 3rd party service.

The following diagram illustrates how it works.

![Serverless architecture](./assets/svls-internal-registry.svg)

1. Build job pushes function image to docker registry using in cluster URL
2. Internal docker registry URL is resolved by K8s DNS to real ip address
3. Kubelet fetch image using URL: `localhost:{node_port}/{image}`
4. NodePort allows kubelet to get into the cluster network and translate `localhost` to `internal-registry.kyma-system.svc.cluster.local` and ask the k8s dns to resolve the name
5. K8s DNS service resolve the name and provide the IP of `internal docker registry`

**NOTE:** Kubelet cannot resolve in cluster URL that's why serverless uses NodePort service.
