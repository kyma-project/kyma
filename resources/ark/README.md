# Ark

## Overview

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Ark documentation](https://heptio.github.io/ark/v0.9.0/).

## Details

By default, Ark comes with GCP as a backup storage provider and no bucket set. With that configuration, the Ark server deployment scales down to 0 replicas because Ark cannot start without the proper configuration for the backup storage bucket. You can change this by providing proper credentials in the heptio-ark/ark secret and changing the configuration in config/default.

