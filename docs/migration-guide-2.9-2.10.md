---
title: Migration Guide 2.9-2.10
---

Due to the removal of `kyma-integration` namespace, we need to migrate all Application Connector secrets to the `kyma-system` namespace. After upgrading the Kyma to the 2.10 version, you need to execute the script:

For Kyma SKR - 
[2.9-2.10-SKR-copy-secrets-to-system-namespace.sh](./assets/2.9-2.10-SKR-copy-secrets-to-system-namespace.sh)

For Kyma Open Source - [2.9-2.10-OS-copy-secrets-to-system-namespace.sh](./assets/2.9-2.10-OS-copy-secrets-to-system-namespace.sh)
