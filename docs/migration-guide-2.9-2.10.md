---
title: Migration Guide 2.9-2.10
---

Due to the removal of the `kyma-integration` Namespace, we need to migrate all Application Connector Secrets to the `kyma-system` Namespace. After upgrading Kyma to version 2.10, you need to execute the following script [2.9-2.10-OS-copy-secrets-to-system-namespace.sh](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.9-2.10-OS-copy-secrets-to-system-namespace.sh).

