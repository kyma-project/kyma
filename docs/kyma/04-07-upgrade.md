---
title: Upgrade Kyma
type: Installation
---

>**CAUTION:** Before you upgrade your Kyma deployment to a newer version, check the release notes of the target release for migration guides. If the target release comes with a migration guide, make sure to follow it closely. If you upgrade to a newer release without performing the steps described in the migration guide, you can compromise the functionality of your cluster or make it unusable altogether.

Upgrading Kyma is the process of migrating from one version of the software to a newer release. This operation depends on release artifacts of target release and migration guides delivered with the target release. 

To upgrade to a version that is several releases newer than the version you're currently using, you must move up to the desired release incrementally. You can skip patch releases.

For example, if you're running Kyma 1.0 and you want to upgrade to version 1.3, you must perform these operations:

1. Upgrade from version 1.0 to version 1.1.
2. Upgrade from version 1.1 to version 1.2.
3. Upgrade from version 1.2 to version 1.3.

>**NOTE:** Kyma does not support a dedicated downgrade procedure. You can achieve a similar result by creating a backup of your cluster before upgrade. Read the [tutorial](/components/backup/#tutorials-taking-backup-using-velero) to learn more about backups.

The upgrade procedure relies heavily on Helm. As a result, the availability of cluster services during the upgrade is not defined by Kyma and can vary from version to version. The existing custom resources (CRs) remain in the cluster.

For more details, read about the [technical aspects](https://github.com/kyma-project/kyma/blob/master/components/kyma-operator/README.md#upgrade-kyma) of the upgrade.

## Upgrade Kyma to a newer version

Follow these steps:

1. Check which version you're currently running. Run this command:
  ```
  kubectl -n kyma-installer get deploy kyma-installer -o jsonpath='{.spec.template.spec.containers[].image}'
  ```
2. Perform the required actions described in the migration guide published with the release you want to upgrade to. Migration guides are linked in the [release notes](https://kyma-project.io/blog/) and are available on the respective [release branches](https://github.com/kyma-project/kyma/branches) in the `docs/migration-guides` directory.
  >**NOTE:** Not all releases require you to perform additional migration steps. If your target release doesn't come with a migration guide, proceed to the next step.
3. Delete the existing `kyma-installer` deployment.
   ```bash
   kubectl delete deployment kyma-installer -n kyma-installer

   ``` 
4. Trigger the upgrade:

    <div tabs name="trigger-the-upgrade" group="upgrade-kyma">
      <details>
      <summary label="local-deployment">
      Local deployment
      </summary>

      - Download the `kyma-config-local.yaml` artifact using following command. Replace `RELEASE_VERSION` with desired version.

      ```
      curl -LO https://storage.googleapis.com/kyma-prow-artifacts/{RELEASE_VERSION}/kyma-config-local.yaml
      ```

      Run this command to apply the overrides required by the new release to your Minikube cluster:

      ```
      kubectl apply -f {KYMA-CONFIG-LOCAL-FILE}
      ```

      >**NOTE:** If you customized your deployment and its overrides, download the `kyma-config-local.yaml` artifact and compare your changes to the overrides of the target release. Merge your changes if necessary.

      - Download the `kyma-installer.yaml` artifact using following command. Replace `RELEASE_VERSION` with desired version:
      ```
      https://storage.googleapis.com/kyma-prow-artifacts/{KYMA_RELEASE}/installer.yaml
      ```

      Run this command to apply the installer file to your Minikube cluster:
      
      ```
      kubectl apply -f {INSTALLER-FILE}
      ```

      - Download the `kyma-installer-cr.yaml` artifact using following command. Replace `RELEASE_VERSION` with desired version.:
      ```
      https://storage.googleapis.com/kyma-prow-artifacts/{KYMA_RELEASE}/kyma-installer-cr.yaml
      ```

      Run this command to apply the installer file to your Minikube cluster:
      
      ```
      kubectl apply -f {INSTALLER-CR-FILE}
      ```

      </details>
      <details>
      <summary label="cluster-deployment">
      Cluster deployment
      </summary>

      >**NOTE:** Before you upgrade a cluster deployment, check if the overrides changed names in the version you're upgrading to.

      - Download the `kyma-installer.yaml` artifact using following command. Replace `RELEASE_VERSION` with desired version:

      ```
      https://storage.googleapis.com/kyma-prow-artifacts/{KYMA_RELEASE}/installer.yaml
      ```

      Run this command to apply the installer file to your cluster:
      
      ```
      kubectl apply -f {INSTALLER-FILE}
      ```

      - Download the `kyma-installer-cr-cluster.yaml` artifact using following command. Replace `RELEASE_VERSION` with desired version.:
      ```
      https://storage.googleapis.com/kyma-prow-artifacts/{KYMA_RELEASE}/kyma-installer-cr-cluster.yaml
      ```

      Run this command to apply the installer file to your cluster:
      
      ```
      kubectl apply -f {KYMA-INSTALLER-CR-CLUSTER-FILE}
      ```

      </details>
    </div>

6. Applying the release artifacts to the cluster triggers the installation of the desired Kyma version. To watch the installation status, run:

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
