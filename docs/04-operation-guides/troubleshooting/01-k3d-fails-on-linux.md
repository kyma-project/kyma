# Provisioning k3d Fails on a Linux Machine {docsify-ignore-all}

## Symptom

You're on a Linux machine and provisioning k3d fails with a message like `Cannot bind to reserved port 80`.

## Cause

By default, provisioning tries to use ports `80` and `433`.
On Linux, the ports are reserved to be used by a privileged user.

## Remedy

Use a custom port for the load balancer. For example, use the port `8080`:

```bash
k3d cluster create -p 8080:80@loadbalancer -p 8443:443@loadbalancer
```

Alternatively, execute the `k3d cluster create` command with sudo privileges.
