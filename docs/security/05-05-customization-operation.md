---
title: OAuth2 server customization and operations
type: Configuration
---

## Credentials backup

The `ory-hydra-credentials` Secret stores all the crucial data required to establish a connection with your database. Nevertheless, it is regenerated every time the ORY chart is upgraded and you may accidentally overwrite your credentials. For this reason, it is recommended to backup the Secret. Run this command to save the contents of the Secret to a file:

```bash
kubectl get secret -n kyma-system ory-hydra-credentials -o yaml > ory-hydra-credentials-$(date +%Y%m%d).yaml
```

## Postgres password update
If Hydra is installed with default setting, a postgres based database will be provided out-of-the-box. If no password has been specified, one is generated and set for the hydra user. This behavior may not always be desired, and in some cases you may want to modify this password. 

In order to set a custom password, provide the override `.Values.global.postgresql.postgresqlPassword` during installation.

In order to update the password for an existing installation, provide the override `.Values.global.postgresql.postgresqlPassword` and perform the update procedure. This however will only change the environmental setting for the database and will not modify the internal database data. In order to update the password in the database please refer to the [postgres documentation](https://www.postgresql.org/docs/11/sql-alteruser.html)
