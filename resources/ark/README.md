# Ark

## Overview

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up and can set TTL (time to live) for stored backups. For more details please see official [Ark documentation](https://heptio.github.io/ark/v0.9.0/).

## Details

By default Ark is installed with GCP as backup storage proivider and no bucket set. In such case the Ark server deployment will be scaled down to 0 replicas (Ark cannot start without proper configuration for backup storage bucket). It can be changed by providing proper credentials in `heptio-ark/ark` secret and configuration in `config/default`.

