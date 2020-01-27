---
title: Upgrade Kyma
type: Installation
---

>**CAUTION:** Before you upgrade your Kyma deployment to a newer version, check the release notes of the target release for migration guides. If the target release comes with a migration guide, make sure to follow it closely. If you upgrade to a newer release without performing the steps described in the migration guide, you can compromise the functionality of your cluster or make it unusable altogether.

Upgrading Kyma is the process of migrating from one version of the software to a newer release. This operation depends on [release artifacts](https://github.com/kyma-project/kyma/releases) listed in the **Assets** section of the GitHub releases page and migration guides delivered with the target release.

To upgrade to a version that is several releases newer than the version you're currently using, you must move up to the desired release incrementally. You can skip patch releases.

For example, if you're running Kyma 1.0 and you want to upgrade to version 1.3, you must perform these operations:

1. Upgrade from version 1.0 to version 1.1.
2. Upgrade from version 1.1 to version 1.2.
3. Upgrade from version 1.2 to version 1.3.

>**NOTE:** Kyma does not support a dedicated downgrade procedure. You can achieve a similar result by restoring your cluster from a backup. Read [this](/components/backup/#overview-overview) document to learn more about backups.

The upgrade procedure relies heavily on Helm. As a result, the availability of cluster services during the upgrade is not defined by Kyma and can vary from version to version. The existing custom resources (CRs) remain in the cluster.

To learn more about the technical aspects of the upgrade, read [this](https://github.com/kyma-project/kyma/blob/master/components/kyma-operator/README.md#upgrade-kyma) document.

## Upgrade Kyma to a newer version

Follow these steps:

1. Check which version you're currently running. Run this command:
  ```
  kubectl -n kyma-installer get deploy kyma-installer -o jsonpath='{.spec.template.spec.containers[].image}'
  ```
2. Perform the required actions described in the migration guide published with the release you want to upgrade to. Migration guides are linked in the [release notes](https://kyma-project.io/blog/) and are available on the respective [release branches](https://github.com/kyma-project/kyma/branches) in the `docs/migration-guides` directory.
  >**NOTE:** Not all releases require you to perform additional migration steps. If your target release doesn't come with a migration guide, proceed to the next step.

3. Trigger the upgrade:

    <div tabs name="trigger-the-upgrade" group="upgrade-kyma">
      <details>
      <summary label="local-deployment">
      Local deployment
      </summary>

      - Download the `kyma-config-local.yaml` artifact. Run this command to apply the overrides required by the new release to your Minikube cluster:
      ```
      kubectl apply -f {KYMA-CONFIG-LOCAL-FILE}
      ```

      >**NOTE:** If you customized your deployment and its overrides, download the `kyma-config-local.yaml` artifact and compare your changes to the overrides of the target release. Merge your changes if necessary.

      - Download the `kyma-installer-local.yaml` artifact and apply it to the cluster to upgrade Kyma. Run:
      ```
      kubectl apply -f {KYMA-INSTALLER-LOCAL-FILE}
      ```

      </details>
      <details>
      <summary label="cluster-deployment">
      Cluster deployment
      </summary>

      >**NOTE:** Before you upgrade a cluster deployment, check if the overrides changed names in the version you're upgrading to.

      Download the `kyma-installer-cluster.yaml` artifact and apply it to the cluster to upgrade Kyma. Run:

      ```
      kubectl apply -f {KYMA-INSTALLER-CLUSTER-FILE}
      ```

      </details>
    </div>

4. Applying the release artifacts to the cluster triggers the installation of the desired Kyma version. To watch the installation status, run:

    <div tabs name="installation-status" group="upgrade-kyma">
      <details>
      <summary label="local-deployment">
      Local deployment
      </summary>

      ```
      ./installation/scripts/is-installed.sh
      ```

      </details>
      <details>
      <summary label="cluster-deployment">
      Cluster deployment
      </summary>

      ```
      while true; do \
      kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
      sleep 5; \
      done
      ```

      </details>
    </div>
