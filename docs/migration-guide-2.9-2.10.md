---
title: Migration Guide 2.9-2.10
---

Due to the removal of the `kyma-integration` Namespace, we need to migrate all Application Connector Secrets to the `kyma-system` Namespace. After upgrading Kyma to the 2.10 version, you will need to execute the script [2.9-2.10-copy-secrets-to-system-namespace.sh](./assets/2.9-2.10-copy-secrets-to-system-namespace.sh). This applies to both the OS and SKR versions. 