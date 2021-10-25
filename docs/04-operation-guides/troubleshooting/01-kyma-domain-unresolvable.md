---
title: Kyma domain is not resolvable
---

## Condition

You get the following error message:

>"The configured Kyma domain {DOMAIN} is not resolvable. This could be due to activated rebind protection of your DNS resolver. Please add virtual service domains to your hosts file."

## Cause

The k3s deployment uses `*.local.kyma.dev` for services that are directed to `127.0.0.1`. If DNS rebind protection is active, resolving this domain fails.

This usually happens when the workstation is using certain routers.

## Remedy

Add the virtual service domains to the host file of your local system.

1. Run the following command:

   ```bash
   sudo kyma import hostfiles 
   ```

2. If that command fails, you get a list of host files.
   Execute the following command, replacing the placeholder with the displayed hosts:
   - For Mac/Linux, run `sudo  /bin/sh -c 'echo \"127.0.0.1 {DISPLAYED_HOSTS}" >> /etc/hosts'`.
   - For Windows, run `echo {DISPLAYED_HOSTS} >> "C:\\Windows\\system32\\drivers\\etc\\hosts"`.
