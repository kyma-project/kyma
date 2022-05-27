---
title: Migration Guide 2.2-2.3
---

Due to the removal of the deprecated Application Connectivity components (such as Application Registry, Connector Service, and Connection Token Handler), some obsolete resources must be deleted.
When you upgrade from Kyma 2.2 to 2.3, either run the script [`2.2-2.3-cleanup-app-connector-resources.sh`](https://github.com/kyma-project/kyma/blob/release-2.3/docs/assets/2.2-2.3-cleanup-app-connector-resources.sh) found in [`/assets`](https://github.com/kyma-project/kyma/tree/release-2.3/docs/assets) or perform the required steps from that script manually.