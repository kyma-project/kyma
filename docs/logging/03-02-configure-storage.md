---
title: Storage configuration examples
type: Details
---

## Storage

By default, Loki comes with the [boltDB](https://github.com/boltdb/bolt) storage configuration. It includes label and index storage, and the filesystem for object storage. Additionally, Loki supports other object stores, such as S3 or GCS.

This is an example of Loki configuration using boltDB and filesystem storage:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: loki
  name: loki
data:
  loki.yaml: |
    auth_enabled: false

    server:
      http_listen_port: 3100

    ingester:
      lifecycler:
        ring:
          store: inmemory
          replication_factor: 1
    schema_config:
      configs:
      - from: 0
        store: boltdb
        object_store: filesystem
        schema: v9
        index:
          prefix: index_
          period: 168h
    storage_config:
      - name: boltdb
        directory: /tmp/loki/index
      - name: filesystem
        directory: /tmp/loki/chunks

```

A sample configuration for GCS looks as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: loki
  name: loki
data:
  loki.yaml: |
    auth_enabled: false

    server:
      http_listen_port: 3100

    ingester:
      lifecycler:
        ring:
          store: inmemory
          replication_factor: 1
    schema_config:
      configs:
      - from: 0
        store: gcs
        object_store: gsc
        schema: v9
        index:
          prefix: index_
          period: 168h
    storage_config:
      gcs:
        bucket_name: <YOUR_GCS_BUCKETNAME>
        project: <BIG_TABLE_PROJECT_ID>
        instance: <BIG_TABLE_INSTANCE_ID>
        grpc_client_config: <YOUR_CLIENT_SETTINGS>
```
