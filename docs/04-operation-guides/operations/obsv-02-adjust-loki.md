---
title: Adjust Loki log limits
---

There's a fixed logs retention time and size. If the default time is exceeded, the oldest logs are removed first.

To adjust the limits to your needs, you simply create a custom YAML file based on the [Loki values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml) and deploy it with `kyma deploy --values-file {VALUES_FILE_PATH}`.

## Adjust log retention period

You can increase the log retention period because you want to see older logs - or decrease it, if you're hitting the volume limits.

In your custom [Loki values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml) file, enter the desired retention period in hours as value for **chunk_store_config.max_look_back_period** and **table_manager.retention_period**.

## Adjust volume size

The way to change the volume size depends on whether you want to do it during the installation or later.

If you know already when installing Kyma that the default storage size isn't going to be enough, expand the Loki volume size for the initial deployment:
In your custom [Loki values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml) file, enter the desired volume size as the value for **persistence.size**.

If you run out of volume later on (for example, because you increased the log retention period), you can increase the volume size by [expanding the Persistent Volume Claims (PVC)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims) for the Loki Pod.
After you edit the PVC, stop the workload and restart it so that the changes take effect.

## Adjust ingestion limit

If your logs persistently exceed the ingestion limit, you can increase that, too.
You don't need to increase the ingestion limit if that happens just occasionally, because log data that couldn't be sent once will go through with the automatic retries in Fluent Bit.

1. In your custom [Loki values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml) file, enter the desired ingestion limit as value for **ingestion_rate_mb**.
1. We recommend that you adjust the values for the **cpu** and **memory** resources.
