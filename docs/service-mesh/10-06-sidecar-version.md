---
title: Istio sidecar version after Kyma upgrade
type: Troubleshooting
---


Kyma has sidecar injection enabled by default - a sidecar is injected to every Deployment in a cluster, without the need to add any labels. For more information, read [this document](#details-sidecar-proxy-injection).

The sidecar version in Pods must match the installed Istio version, otherwise mesh connectivity may be broken.
This is not an issue when installing Kyma, but it may be a problem during upgrades. When Kyma is upgraded to a new version along with a new Istio version, existing sidecars injected into Pods remain in an original version.
Kyma contains an `istio-proxy-reset` job, that performs a rollout for most common workload types like deployments, daemonsets, etc. The job ensures all Kyma components are properly updated.
Certain kinds of user-defined workloads can't be rolled out automatically, for example, a standalone Pod without any backing management mechanism (like ReplicaSet or a Job).
Such user-defined workloads (that are not part of Kyma) must be manually restarted to work correctly with the updated Istio version.

To find if any pods/workloads require a manual restart, you can:

* Get the list of Pods with outdated sidecar version. You can get the Istio version for a Kyma release from the Istio chart version located in `resources/Istio/Chart.yaml` file. The returned list is in `name/namespace` format, empty output means no Pods require migration. Use the following command:

    ```bash
    EXPECTED_ISTIO_PROXY_IMAGE="1.7.4-distroless" #Valid for Kyma: 1.17.x, 1.18.x
    COMMON_ISTIO_PROXY_IMAGE_PREFIX="eu.gcr.io/kyma-project/external/istio/proxyv2"
    kubectl get pods -A -o json | jq -rc '.items | .[] | select(.spec.containers[].image | startswith("'"${COMMON_ISTIO_PROXY_IMAGE_PREFIX}"'") and (endswith("'"${EXPECTED_ISTIO_PROXY_IMAGE}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'
    ```


* Run the `istio-proxy-reset` script in dry-run mode. Execute the command from within the directory with checked-out Kyma sources. You can get the Istio version for the Kyma release from the Istio chart version located in `resources/Istio/Chart.yaml` file. The output contains information about objects (Pods, Deployments, etc.) that require rollout.

    ```bash
    EXPECTED_ISTIO_PROXY_IMAGE="1.7.4-distroless" #Valid for Kyma: 1.17.x, 1.18.x
    COMMON_ISTIO_PROXY_IMAGE_PREFIX="eu.gcr.io/kyma-project/external/istio/proxyv2"
    DRY_RUN="true"
    ./resources/istio/files/istio-proxy-reset.sh
    ```

After you found a set of objects that require the manual update, restart their related workloads so that new Istio sidecars are injected into the Pods.
