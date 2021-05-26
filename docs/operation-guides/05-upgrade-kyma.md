---
title: Upgrade Kyma
---

The `deploy` command not only installs Kyma, you also use it to upgrade the Kyma version on the cluster. For example, you can specify the sources, components, and configuration values you want to use. For code examples, see [Custom Kyma Installation](#install-kyma-custom).

> **NOTE:** If you upgrade from one Kyma release to a newer one, an automatic compatibility check compares your current version and the new release.<br>
The compatibility check only works with release versions. If you installed Kyma from a PR, branch, revision, or local version, the compatibility cannot be checked.