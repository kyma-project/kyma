---
title: Migration Guide 2.9-2.10
---

Due to the removal of the `kyma-integration` Namespace, we need to migrate all Application Connector Secrets to the `kyma-system` Namespace. After upgrading the Kyma to the 2.10 version, you need to execute the following script [2.9-2.10-OS-copy-secrets-to-system-namespace.sh](./assets/2.9-2.10-OS-copy-secrets-to-system-namespace.sh)

