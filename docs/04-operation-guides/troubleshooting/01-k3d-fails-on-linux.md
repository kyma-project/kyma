---
title: Provisioning k3d Fails on a Linux Machine
---

## Symptom

You're on a Linux machine and provisioning k3d fails with a message like `Cannot bind to reserved port 80`.

## Cause

By default, provisioning tries to use ports `80` and `433`.
On Linux, the ports are reserved to be used by a privileged user.

## Remedy

Use a custom port for the load balancer. For example, use the port `8080`:
```bash
kyma provision k3d -p 8080:80@loadbalancer -p 8443:443@loadbalancer
```

Alternatively, execute the `kyma provision k3d` command with sudo privileges.
