---
title: Overview
---

Backup in Kyma uses [Ark](https://github.com/heptio/velero/) (The Project is renamed to Velero. As soon as there is a new Velero Version, Ark will be called Velero). Ark is a tool to backup kubernetes resources and store them in a Blob storage. Beside that, Ark is triggering Physical Volumen Snapshots and stores the snapshot references as part of the Backup.

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups.