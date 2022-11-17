---
title: Migration Guide 2.8-2.9
---

Due to the [deprecation of PodSecurityPolicy](https://kubernetes.io/blog/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/) with Kubernetes 1.21 and the plan of removal in the 1.25 release, we removed the usage of PSPs for many of our Kyma resources. To delete left-over PSP resources, when you upgrade from Kyma 2.8 to 2.9, either run the script [2.8-2.9-cleanup-psp.sh](./assets/2.8-2.9-cleanup-psp.sh) or run the commands from the script manually.