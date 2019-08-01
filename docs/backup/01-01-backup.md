---
title: Overview
---

Kyma integrates with [Velero](https://github.com/heptio/velero/) to provide backup and restore capabilities.

Velero backs up Kubernetes resources and stores them in buckets of [supported cloud providers](https://velero.io/docs/v1.0.0/support-matrix/). It triggers physical volume snapshots and includes the snapshot references in the backup. Velero can create scheduled or on-demand backups, filter objects to include in the backup, and set time to live (TTL) for stored backups.

If you provide the needed configuration for the Velero installation as explained [here](/components/backup/#installation-installation), Kyma by default sets up a scheduled backup that runs once a day every day from Monday to Friday. It takes a backup using the provided configurations. To change this behavior or its configurations, change **schedules** configuration. See more in this [guide](/components/backup/#configuration-velero-chart).

For more details, see the official [Velero documentation](https://velero.io/docs/v1.0.0).
