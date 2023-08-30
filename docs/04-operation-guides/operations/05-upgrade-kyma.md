---
title: Upgrade or Downgrade Kyma
---

## Upgrade

You can use the `deploy` command not only to install Kyma, but also to upgrade the Kyma version on the cluster. For example, you can specify the sources, components, and configuration values you want to use.

For code examples, see [Install Kyma](02-install-kyma.md).

> **NOTE:** If you upgrade from one Kyma release to a newer one, the automatic compatibility check compares your current version and the new release.<br>
The compatibility check only works with release versions. If you installed Kyma from a PR, branch, revision, or local version, the compatibility cannot be checked.

> **CAUTION:** Zero-downtime upgrades are not supported.

## Downgrade

Downgrading your Kyma version is not supported.

> **NOTE:** To learn how to upgrade a Kyma module go to [Install, uninstall and upgrade a Kyma module](../../02-get-started/08-enable-disable-upgrade-kyma-module.md#upgrade-a-kyma-module).
