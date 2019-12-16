---
title: Charts
type: Details
---

Kyma uses Helm charts to deliver single components and extensions, as well as the core components. This document contains information about the chart-related technical concepts, dependency management to use with Helm charts, and chart examples.

## Manage dependencies with Init Containers

The **ADR 003: Init Containers for dependency management** document declares the use of Init Containers as the primary dependency mechanism.

[Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) present a set of distinctive behaviors:

* They always run to completion.
* They start sequentially, only after the preceding Init Container completes successfully.
  If any of the Init Containers fails, the Pod restarts. This is always true, unless the `restartPolicy` equals `never`.

[Readiness Probes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes) ensure that the essential containers are ready to handle requests before you expose them. At a minimum, probes are defined for every container accessible from outside of the Pod. It is recommended to pair the Init Containers with readiness probes to provide a basic dependency management solution.

## Examples

Here are some examples:

1. Generic

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: nginx-deployment
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: nginx
     template:
       metadata:
         labels:
           app: nginx
       spec:
         containers:
         - name: nginx
           image: nginx:1.7.9
           ports:
           - containerPort: 80
           readinessProbe:
             httpGet:
               path: /healthz
               port: 80
             initialDelaySeconds: 30
             timeoutSeconds: 1
   ```

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: myapp-pod
   spec:
     initContainers:
     - name: init-myservice
       image: busybox
       command: ['sh', '-c', 'until nslookup nginx; do echo waiting for nginx; sleep 2; done;']
     containers:
     - name: myapp-container
       image: busybox
       command: ['sh', '-c', 'echo The app is running! && sleep 3600']
   ```

2. Kyma

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: helm-broker
     labels:
       app: helm-broker
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: helm-broker
     strategy:
       type: RollingUpdate
       rollingUpdate:
         maxUnavailable: 0
     template:
       metadata:
         labels:
           app: helm-broker
       spec:

         initContainers:
         - name: init-helm-broker
           image: eu.gcr.io/kyma-project/alpine-net:0.2.74
           command: ['sh', '-c', 'until nc -zv service-catalog-controller-manager.kyma-system.svc.cluster.local 8080; do echo waiting for etcd service; sleep 2; done;']

         containers:
         - name: helm-broker
           ports:
           - containerPort: 6699
           readinessProbe:
             tcpSocket:
               port: 6699
             failureThreshold: 3
             initialDelaySeconds: 10
             periodSeconds: 3
             successThreshold: 1
             timeoutSeconds: 2
   ```

## Support for the Helm wait flag

High level Kyma components, such as **core**, come as Helm charts. These charts are installed as part of a single Helm release. To provide ordering for these core components, the Helm client runs with the `--wait` flag. As a result, Tiller waits for the readiness of all of the components, and then evaluates the readiness.

For `Deployments`, set the strategy to `RollingUpdate` and set the `MaxUnavailable` value to a number lower than the number of replicas. This setting is necessary, as readiness in Helm v2.10.0 is fulfilled if the number of replicas in ready state is not lower than the expected number of replicas:

```
ReadyReplicas >= TotalReplicas - MaxUnavailable
```

## Chart installation details

The Tiller server performs the chart installation process. This is the order of operations that happen during the chart installation:

* resolve values
* recursively gather all templates with the corresponding values
* sort all templates
* render all templates
* separate hooks and manifests from files into sorted lists
* aggregate all valid manifests from all sub-charts into a single manifest file
* execute PreInstall hooks
* create a release using the ReleaseModule API and, if requested, wait for the actual readiness of the resources
* execute PostInstall hooks

## Notes

All notes are based on Helm v2.10.0 implementation and are subject to change in future releases.

* Regardless of how complex a chart is, and regardless of the number of sub-charts it references or consists of, it's always evaluated as one. This means that each Helm release is compiled into a single Kubernetes manifest file when applied on API server.

* Hooks are parsed in the same order as manifest files and returned as a single, global list for the entire chart. For each hook the weight is calculated as a part of this sort.

* Manifests are sorted by `Kind`. You can find the list and the order of the resources on the Kubernetes [Tiller](https://github.com/kubernetes/helm/blob/v2.10.0/pkg/tiller/kind_sorter.go#L29) website.

## Glossary

* **resource** is any document in a chart recognized by Helm or Tiller. This includes manifests, hooks, and notes.
* **template** is a valid Go template. Many of the resources are also Go templates.
