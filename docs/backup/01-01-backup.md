---
title: Overview
---

Kyma integrates with [Velero](https://github.com/heptio/velero/) to provide backup and restore capabilities.

Velero backs up Kubernetes resources and stores them in buckets of [supported cloud providers](https://velero.io/docs/v1.2.0/supported-providers/). It triggers physical volume snapshots and includes the snapshot references in the backup. Velero can create scheduled or on-demand backups, filter objects to include in the backup, and set time to live (TTL) for stored backups.

If you configured Velero when installing Kyma as explained [here](/components/backup/#installation-installation), backup is enabled with the default schedule and runs once a day every day from Monday to Friday. To change the settings of the backup change the **schedules** configuration in the Velero chart [configuration](/components/backup/#configuration-velero-chart). Enabling out-of-box scheduled backups is the main reason for bundling Velero within Kyma installation.

For more details, see the official [Velero documentation](https://velero.io/docs/v1.2.0).
