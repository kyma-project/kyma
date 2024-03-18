# Kyma Domain Is Not Resolvable

## Condition

You get the following error message:

```bash
The configured Kyma domain {DOMAIN} is not resolvable. This could be due to activated rebind protection of your DNS resolver. Please add virtual service domains to your hosts file."
```

## Cause

The k3d deployment uses `*.local.kyma.dev` for services that are directed to `127.0.0.1`. If DNS rebind protection is active, resolving this domain fails.

This usually happens when the workstation is using certain routers.

## Remedy

Add the VirtualService domains manually to the hosts file of your local system.

Run the following command:

```bash
sudo kyma import hosts 
```
