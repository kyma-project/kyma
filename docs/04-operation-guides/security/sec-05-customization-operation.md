---
title: OAuth2 server customization and operations
---

## Credentials backup

The `ory-hydra-credentials` Secret stores all the crucial data required to establish a connection with your database. Nevertheless, it is regenerated every time the ORY chart is upgraded and you may accidentally overwrite your credentials. For this reason, it is recommended to backup the Secret. Run this command to save the contents of the Secret to a file:

```bash
kubectl get secret -n kyma-system ory-hydra-credentials -o yaml > ory-hydra-credentials-$(date +%Y%m%d).yaml
```

## Postgres password update

If Hydra is installed with the default settings, a Postgres-based database is provided out-of-the-box. If no password was specified, one is generated and set for the Hydra user. This behavior may not always be desired, so in some cases you may want to modify this password.

In order to set a custom password, provide the `.Values.global.postgresql.postgresqlPassword` override during installation.

In order to update the password for an existing installation, provide the `.Values.global.postgresql.postgresqlPassword` override and perform the update procedure. However, this only changes the environmental setting for the database and does not modify the internal database data. In order to update the password in the database, please refer to the [Postgres documentation](https://www.postgresql.org/docs/11/sql-alteruser.html).
