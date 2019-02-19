---
title: Overview
---

Logging in Kyma uses [Loki](https://github.com/grafana/loki) which is a Prometheus-like log management system. This lightweight solution, integrated with Grafana, is easy to understand and operate. Loki provides a log router (promtail) for Docker containers. The router runs inside Docker, connects with each container on a host and routes the container's logs to the log management system.
