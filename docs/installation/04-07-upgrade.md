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

>**NOTE:** Kyma does not support a dedicated downgrade procedure. You can achieve a similar result by creating a backup of your cluster before upgrade. Read the [tutorial](/components/backup/#tutorials-taking-backup-using-velero) to learn more about backups.

The upgrade procedure relies heavily on Helm. As a result, the availability of cluster services during the upgrade is not defined by Kyma and can vary from version to version. The existing custom resources (CRs) remain in the cluster.

For more details, read about the [technical aspects](https://github.com/kyma-project/kyma/blob/master/components/kyma-operator/README.md#upgrade-kyma) of the upgrade.

## Upgrade Kyma to a newer version

Follow these steps:

1. Kyma CLI should be in the same version of the release you would like to upgrade to. To check which version you're currently running, run this command:

  ```bash
  kyma version
  ```

2. Perform the required actions described in the migration guide published with the release you want to upgrade to. Migration guides are linked in the [release notes](https://kyma-project.io/blog/) and are available on the respective [release branches](https://github.com/kyma-project/kyma/branches) in the `docs/migration-guides` directory.
  >**NOTE:** Not all releases require you to perform additional migration steps. If your target release doesn't come with a migration guide, proceed to the next step.

3. Trigger the upgrade:
  >**CAUTION:** Do not forget to supply the same overrides using the `-o` flag and the same component list using the `-c` flag if you provided any of them during the installation. There might be new components on the version that you would like to upgrade to. It is important to add them also to your custom component list.

  ```bash
  kyma upgrade -s {VERSION}
  ```

  If you want to upgrade Kyma to use one of the predefined [profiles](#installation-overview-profiles), run:

  ```bash
  kyma upgrade -s {VERSION} --profile {evaluation|production}
  ```
