---
title: Kyma domain is not resolvable
---

## Condition

When you {were doing something}, you get the following error message:

>"The configured Kyma domain {DOMAIN} is not resolvable. This could be due to activated rebind protection of your DNS resolver. Please add virtual service domains to your hosts file."

## Cause

This error is specific to a few brands of routers. {Expand explanation, if necessary}

## Remedy

Add the virtual service domains to the host file of your local system.

Run the following command:

```bash
kyma store hostfiles 
```
