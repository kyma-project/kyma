---
title: Acquiring NATS server system account credentials
---

To acquire the username and password use:
```bash
kubectl get secrets -n kyma-system eventing-nats-secret -oyaml | grep -e accountsJson | awk '{print $2}' | base64 -d | grep {user: | awk '{$1=$1};1' | awk '{print substr($0, 2, length($0) - 2)}'
```
These credentials can be used to access the aforementioned NATS resources with the cli tool by passing the `--user admin` and `--password <your password>` options, e. g.:
```bash
nats server info --user admin --password <your password>
```

---
title: {Document title}
---

<!-- Use this template to write "how-to" instructions that enable users to accomplish a task. Each task topic should tell how to perform a single, specific procedure.

You can use this template for any step-by-step instruction, no matter whether it's a task during the getting started guide, a tutorial for software developers, or an operational guide.

For the document file name, follow the pattern `{COMPONENT_ABBRV}-{NUMBER_PER_COMPONENT}-{FILE_NAME}.md`.

Select a title that describes the task that's accomplished, not the documented software feature. For example, use "Define resource consumption", not "Select a profile". Use the imperative "Select...", rather than gerund form "Selecting..." or "How to select...".

With regards to structure, it’s nice to have an **introductory paragraph** ("why would I want to do this task?"), **prerequisites** if needed, then the **steps**, and finally the expected **result** that shows the operation was successful.
It's good practice to have 5-9 steps; anything longer can probably be split.
-->

## Context

Accessing certain resources in NATS requires [`system_account` privileges](https://docs.nats.io/running-a-nats-service/configuration/sys_accounts). Kyma will automatically generate a `system account` user via a Secret dubbed `eventing-nats-secret` in the `kyma-system` Namespace.

## Prerequisites

None.

## Procedure

To acquire the cretentials run the following command:

```bash
kubectl get secrets -n kyma-system eventing-nats-secret -oyaml | grep -e accountsJson | awk '{print $2}' | base64 -d | grep {user: | awk '{$1=$1};1' | awk '{print substr($0, 2, length($0) - 2)}'
```

This will return the credentials for the `system account` user.
```bash
user: admin, password: <your password>
```

To access the NATS server with the [nats-cli tool](https://github.com/nats-io/natscli) you need to first forward its port:
```bash
kubectl port-forward -n kyma-system svc/eventing-nats 4222
```
Now you can send your nats commands by passing the credentials:
```bash
nats server info --user admin --password <your password>
```
