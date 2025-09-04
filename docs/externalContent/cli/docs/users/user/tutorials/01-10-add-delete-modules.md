# Adding and Deleting a Kyma Module Using Kyma CLI

This tutorial shows how you can add and delete a new module using Kyma CLI.

> [!WARNING]
> Modules added without any specified form of the custom resource have the policy field set to `Ignore`.

## Procedure

### Adding a New Module

To add a new module with the default policy set to `CreateAndDelete`, use the following command:

```bash
kyma alpha module add {MODULE-NAME} --default-cr
```

To add a module with a different CR, use the `--cr-path={CR-FILEPATH}` flag:

```bash
kyma alpha module add {MODULE-NAME} --cr-path={CR-PATH-FILEPATH}
```

You can also specify which channel the module should use with the `-c {CHANNEL-NAME}` flag:

```bash
kyma alpha module add {MODULE-NAME} -c {CHANNEL-NAME} --default-cr
```

### Deleting an Existing Module

To delete an existing module, use the following command:

```bash
kyma alpha module delete {MODULE-NAME} 
```
