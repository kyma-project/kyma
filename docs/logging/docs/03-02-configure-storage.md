---
title: Storage Configuration
type: Details
---

#### Storage
By default, Loki installation comes with storage configuration [boltdb](https://github.com/boltdb/bolt) including label/index storage and the filesystem for the object store. Additionally, Loki supports other object stores, such as S3 or GCS.

The following configuration shows Loki configuration using boltdb and filesystem:
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
    storage_configs:
      - name: boltdb
        directory: /tmp/loki/index
      - name: filesystem
        directory: /tmp/loki/chunks

``` 

The Loki storage configuration consists of the **schema_config** and **storage_configs** definitions. Use the **schema_config** to define your storage types, and **storage_configs** to configure the already defined storage types.

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
    storage_configs:
      gcs:
        bucket_name: <YOUR_GCS_BUCKETNAME>
        project: <BIG_TABLE_PROJECT_ID>
        instance: <BIG_TABLE_INSTANCE_ID>
        grpc_client_config: <YOUR_CLIENT_SETTINGS>
       
```
