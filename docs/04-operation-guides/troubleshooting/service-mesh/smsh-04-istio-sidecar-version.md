---
title: Incompatible Istio sidecar version after Kyma upgrade
---

## Symptom

You upgraded Kyma and mesh connectivity is broken.

## Cause

By default, Kyma has sidecar injection disabled - there is no automatic sidecar injection into any Pod in a cluster. For more information, read the document about [enabling Istio sidecar proxy injection](../../operations/smsh-01-istio-enable-sidecar-injection.md).

The sidecar version in Pods must match the installed Istio version. Otherwise, mesh connectivity may be broken.
This issue may appear during Kyma upgrade. When Kyma is upgraded to a new version along with a new Istio version, existing sidecars injected into Pods remain in an original version.
Kyma contains `istio-proxy-reset` <!--`istio-proxy-reset` is no longer a job. Update and explain what `istio-proxy-reset` actually is once Reconciller is ready to use.--> that performs a rollout for most common workload types, such as Deployments, DaemonSets, etc. The job ensures all Kyma components are properly updated.
However, some user-defined workloads can't be rolled out automatically. This applies, for example, to a standalone Pod without any backing management mechanism, such as a ReplicaSet or a Job.
Such user-defined workloads, that are not part of Kyma, must be manually restarted to work correctly with the updated Istio version.

## Remedy

To check if any Pods or workloads require a manual restart, follow these steps:

1. Check the installed Istio version using one of these methods:

* From the `istiod` deployment in a running cluster, run:

   ```bash
   export KYMA_ISTIO_VERSION=$(kubectl get deployment istiod -n istio-system -o json | jq '.spec.template.spec.containers | .[].image' | sed 's/[^:"]*[:]//' | sed 's/["]//g')
   ```

* From Kyma sources, run this command from within the directory that contains Kyma sources:

   ```bash
   export KYMA_ISTIO_VERSION=$(cat resources/istio/Chart.yaml | grep version | sed 's/[^:]*[:]//' | sed 's/ //g')
   ```

2. Get the list of objects which require rollout. Find all Pods with outdated sidecars. The returned list follows the `name/namespace` format. The empty output means that there is no Pod that requires migration. To find all outdated Pods, run:

     <!--The command in step 2 can change once we start using solo.io images.-->

   ```bash
   COMMON_ISTIO_PROXY_IMAGE_PREFIX="europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2"
   kubectl get pods -A -o json | jq -rc '.items | .[] | select(.spec.containers[].image | startswith("'"${COMMON_ISTIO_PROXY_IMAGE_PREFIX}"'") and (endswith("'"${KYMA_ISTIO_VERSION}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'
   ```

3. After you find a set of objects that require the manual update, restart their related workloads so that new Istio sidecars are injected into the Pods.
