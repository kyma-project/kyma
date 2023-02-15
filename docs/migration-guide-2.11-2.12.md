---
title: Migration Guide 2.11-2.12
---

In preparation for the upcoming modularization and having a reduced set of dependencies on a module, Kyma switched to an annotation-based scraping of metrics for system components. With that, the ServiceMonitors of the components must be cleaned up. When you upgrade from Kyma 2.11 to 2.12, either run the script [cleanup.sh](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-cleanup-servicemonitors.sh) or run the commands from the script manually.
