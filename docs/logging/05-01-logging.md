---
title: Logging chart
type: Configuration
---

To configure the Logging chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **persistence.enabled** | Specifies whether you store logs on a persistent volume instead of a volatile mounted volume. | `true` |
| **persistence.size** | Defines the size of the persistent volume. | `10Gi` |
| **config.auth_enabled** | Authenticates the tenant sending the request to the logging service when Loki runs in the multi-tenant mode. Setting it to `true` requires authentication using the HTTP (`X-Scope-OrgID`) header. Since Kyma supports the single-tenant mode only, you must set this parameter to `false`. This way, Loki does not require the `X-Scope-OrgID` header and the tenant ID defaults to `fake`. | `false` |
| **config.ingester.lifecycler.address** | Specifies the address of the lifecycler that coordinates distributed logging services. | `127.0.0.1` |
| **config.ingester.lifecycler.ring.store** | Specifies the storage for information on logging data and their copies. | `inmemory` |
| **config.ingester.lifecycler.ring.replication_factor** | Specifies the number of data copies on separate storages. | `1` |
| **config.schema_configs.from** | Specifies the date from which index data is stored. | `0` |
| **config.schema_configs.store** | Specifies the storage type. `boltdb` is an embedded key-value storage that stores the index data. | `boltdb` |
| **config.schema_configs.object_store** | Specifies if you use local or cloud storages for data. | `filesystem` |
| **config.schema_configs.schema** | Defines the schema version that Loki provides. | `v9` |
| **config.schema_configs.index.prefix** | Specifies the prefix added to all index file names to distinguish them from log chunks. | `index_` |
| **config.schema_configs.index.period** | Defines how long indexes and log chunks are retained. | `168h` |
| **config.storage_config.boltdb.directory** | Specifies the physical location of indexes in `boltdb`. | `/data/loki/index` |
| **config.storage_config.filesystem.directory** | Specifies the physical location of log chunks in `filesystem`. | `/data/loki/chunks` |

>**NOTE:** The Loki storage configuration consists of the **schema_config** and **storage_config** definitions. Use **schema_config** to define the storage types and **storage_config** to configure storage types that are already defined.
