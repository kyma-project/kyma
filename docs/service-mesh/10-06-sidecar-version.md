---
title: Incompatible Istio sidecar version after Kyma upgrade
type: Troubleshooting
---


Kyma has sidecar injection enabled by default - a sidecar is injected to every Pod in a cluster without the need to add any labels. For more information, read [this document](#details-sidecar-proxy-injection).

The sidecar version in Pods must match the installed Istio version. Otherwise, mesh connectivity may be broken.
This issue may appear during Kyma upgrade. When Kyma is upgraded to a new version along with a new Istio version, existing sidecars injected into Pods remain in an original version.
Kyma contains the `istio-proxy-reset` job that performs a rollout for most common workload types, such as Deployments, DaemonSets, etc. The job ensures all Kyma components are properly updated.
However, some user-defined workloads can't be rolled out automatically. This applies, for example, to a standalone Pod without any backing management mechanism, such as a ReplicaSet or a Job.
Such user-defined workloads, that are not part of Kyma, must be manually restarted to work correctly with the updated Istio version.

To check if any Pods or workloads require a manual restart, follow these steps:

1. Check the installed Istio version using one of these methods:

    * From the `istiod` deployment in a running cluster:
        ```bash
        export KYMA_ISTIO_VERSION=$(kubectl get deployment istiod -n istio-system -o json | jq '.spec.template.spec.containers | .[].image' | sed 's/[^:"]*[:]//' | sed 's/["]//g')
        ```

    * From Kyma sources - run this command from within the directory that contains Kyma sources:
        ```bash
        export KYMA_ISTIO_VERSION=$(cat resources/istio/Chart.yaml | grep version | sed 's/[^:]*[:]//' | sed 's/ //g')
        ```

2. Get the list of objects which require rollout using one of these methods:

    * Find all Pods with outdated sidecars. The returned list follows the `name/namespace` format. The empty output means that there is no Pod that requires migration. To find all outdated Pods, run:
        ```bash
        COMMON_ISTIO_PROXY_IMAGE_PREFIX="eu.gcr.io/kyma-project/external/istio/proxyv2"
        kubectl get pods -A -o json | jq -rc '.items | .[] | select(.spec.containers[].image | startswith("'"${COMMON_ISTIO_PROXY_IMAGE_PREFIX}"'") and (endswith("'"${KYMA_ISTIO_VERSION}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'
        ```


    * Run the `istio-proxy-reset` script in the dry-run mode. The output contains information about objects, such as Pods, Deployments, etc., that require rollout. To run the script, run this command from within the directory with checked-out Kyma sources:
        ```bash
        EXPECTED_ISTIO_PROXY_IMAGE="${KYMA_ISTIO_VERSION}"
        COMMON_ISTIO_PROXY_IMAGE_PREFIX="eu.gcr.io/kyma-project/external/istio/proxyv2"
        DRY_RUN="true"
        ./resources/istio/files/istio-proxy-reset.sh
        ```

After you find a set of objects that require the manual update, restart their related workloads so that new Istio sidecars are injected into the Pods.
