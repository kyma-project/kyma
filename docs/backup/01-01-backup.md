---
title: Overview
---

The backup in Kyma uses [Velero](https://github.com/heptio/velero/).

Velero backs up Kubernetes resources and stores them in Azure Blob storage. It triggers physical volume snapshots and includes the snapshot references in the backup. Ark can create scheduled or on-demand backups, filter objects to include in the backup, and set time to live (TTL) for stored backups.

For more details, see the official [Velero documentation](https://velero.io/docs/v1.0.0).