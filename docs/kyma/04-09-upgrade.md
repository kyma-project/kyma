---
title: Upgrade Kyma
type: Installation
---

Upgrading Kyma is the process of migrating from one version of the software to a newer release. This operation depends on [release artefacts](https://github.com/kyma-project/kyma/releases) listed in the "assets" section of the GitHub releases page and migration guides delivered with the target release.

To upgrade to a version that is several releases newer than the version you're currently using, you must move up to the desired release incrementally. You can skip patch releases.

For example, ff you're running Kyma 1.0 and you want to upgrade to version 1.3, you must perform these operations:
  1. Upgrade from version 1.0 to version 1.1.
  2. Upgrade from version 1.1 to version 1.2.
  3. Upgrade from version 1.2 to version 1.3.  

>**NOTE:** Kyma does not support a dedicated downgrade procedure. You can achieve a similar result by restoring your cluster from a backup. Read [this](/components/backup/#overview-overview) document to learn more about backups.

## Upgrade your deployment to a newer version

Follow these steps:

1. Check which version you're currently running. Run this command:
  ```
  kubectl -n kyma-installer get deploy kyma-installer -o jsonpath='{.spec.template.spec.containers[].image}
  ```
2. Perform the required actions described in the migration guide published with the release you want to upgrade to. Migration guides are linked in the [release notes](https://kyma-project.io/blog/) and are available on the respective [release branches](https://github.com/kyma-project/kyma/branches) in the `docs/migration-guides` directory.
  >**NOTE:** Not all releases require you to perform additional migration steps. If your target release doesn't come with a migration guide, proceed to the next step.

3. Trigger the upgrade:
    <div tabs>
      <details>
      <summary>
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
      <summary>
      Cluster deployment
      </summary>

      Download the `kyma-installer-cluster.yaml` artifact and apply it to the cluster to upgrade Kyma. Run:

      ```
      kubectl apply -f {KYMA-INSTALLER-CLUSTER-FILE}
      ```

      </details>
    </div>
