---
title: Storage Configuration
---

#### Storage
Loki installation delivered by default storage configuration <b>[boltdb](https://github.com/boltdb/bolt)</b> as label/index storage and <b>filesystem</b> for object store however loki support object stores like S3, GCS etc.

Following configuration shown loki with boltdb and filesystem
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

Loki storage configuration divided in two section <b>schema_config</b> and <b>storage_configs</b>, under schema_config you can define your storage types and under storage_configs specific configuration for declared storage types.

##### Example Configuration for GCS

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