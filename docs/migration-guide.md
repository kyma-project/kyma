---
title: Migration Guide
---

In preparation for the upcoming modularization and having a reduced set of dependencies on a module, Kyma switched to an annotation-based scraping of metrics for system components. With that, the ServiceMonitors of the components need to get cleaned up. When you upgrade from Kyma 2.9 to 2.10, either run the script [cleanup.sh](./assets/cleanup.sh) or run the commands from the script manually.
