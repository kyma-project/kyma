---
title: Architecture
---

This document outlines the logging architecture of Kyma. It highlights information sources from which Logspout extracts logs to feed to OK Log.

Logspout is deployed as a stateless Daemonset and shares `/var/run/docker.sock` from the node. It then feeds the logs to OK Log through the ingest API.

OK Log is deployed as a Statefulset. It is capable for storing the logs for 7 days, which is a configurable property. Read more on OK Log architectural decisions [here](https://github.com/oklog/oklog/tree/master/doc/arch).
