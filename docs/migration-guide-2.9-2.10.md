---
title: Migration Guide 2.9-2.10
---

Due to the removal of `kyma-integration` namespace, we need to migrate all Application Connector secrets to the `kyma-system` namespace. Before upgrading the Kyma version, you need to execute the script [2.9-2.10-copy-secrets-to-system-namespace.sh](./assets/2.9-2.10-copy-secrets-to-system-namespace.sh). This applies to both OS and SKR versions. 