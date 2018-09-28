---
title: Overview
---

Logging in Kyma uses [Logspout](https://github.com/gliderlabs/logspout) and [OK Log](https://github.com/oklog/oklog). Logspout is a log router for Docker containers that runs inside Docker. It attaches to each container on a host and routes their logs to a Log Management System. OK Log is a distributed coordination-free log management system. It is a lightweight solution which is not only easy to understand but also easy to operate.