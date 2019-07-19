---
title: Overview
---

Kyma integrates with [Velero](https://github.com/heptio/velero/) to provide backup and restore capabilities.

Velero backs up Kubernetes resources and stores them in buckets of [supported cloud providers](https://velero.io/docs/v1.0.0/support-matrix/). It triggers physical volume snapshots and includes the snapshot references in the backup. Velero can create scheduled or on-demand backups, filter objects to include in the backup, and set time to live (TTL) for stored backups.

For more details, see the official [Velero documentation](https://velero.io/docs/v1.0.0).
