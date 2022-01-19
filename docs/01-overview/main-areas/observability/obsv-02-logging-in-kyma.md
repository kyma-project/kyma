---
title: Logging in Kyma
---

For logging, Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system. 
By default, Loki stores up to 30 GB of data for a maximum of 5 days, with maximum ingestion rate of 3 MB/s. If the default time is exceeded, the oldest logs are removed first.
