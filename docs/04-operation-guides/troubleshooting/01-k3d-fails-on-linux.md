---
title: Provisioning k3d fails on a Linux machine
---

## Symptom

You're on a Linux machine and provisioning k3d fails with a message like `Cannot bind to reserved port 80`.

## Cause

By default, provisioning tries to use ports 80 and 433.
On Linux, the port is reserved.

## Remedy

Use a custom port for the load balancer, for example, port 8080.
To do this, run `kyma provision k3d -p 8080:8080@loadbalancer -p 8443:8443@loadbalancer`.

Alternatively, execute the `kyma provision` command with sudo privileges.
