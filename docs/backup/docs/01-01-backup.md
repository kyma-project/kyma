---
title: Overview
---

The backup in Kyma uses [Ark](https://github.com/heptio/velero/).

>**NOTE:** The Ark project changed its name to Velero. The full migration is planned once the new Velero version is available.

Ark backs up Kubernetes resources and stores them in Azure Blob storage. It triggers physical volume snapshots and includes the snapshot references in the backup. Ark can create scheduled or on-demand backups, filter objects to include in the backup, and set time to live (TTL) for stored backups.

For more details, see the official [Ark documentation](https://heptio.github.io/velero/v0.9.0/).