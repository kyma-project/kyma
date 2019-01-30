---
title: Overview
---

Logging in Kyma uses [Loki](https://github.com/grafana/loki). Loki provide a log router (promtail) for Docker containers that runs inside Docker. It attaches to each container on a host and routes their logs to a Log Management System. Loki is a prometheus like log management system. It is a lightweight solution which is not only easy to understand but also easy to operate and integrated in Grafana.