---
title: Overview
type: Installation
---
[todo]

Kyma is a complex tool which consists of many different [components](#details-details) that provide various functionalities to extend your application.

You can install either from the release or from sources. You can also install... [todo]


Due to this fact, we provide Kyma modularity , you can choose components you want to include in the Kyma installation.
These components do not install with Kyma Lite:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the SQL Server. |
| `port` | `int	` | The port number to connect to on the SQL Server. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |

Follow these installation guides to install Kyma locally:
* Install Kyma locally from release
* Install Kyma locally from sources

You can also install Kyma depending on the supported cloud providers:
- Install Kyma on GKE cluster
- Install Kyma on AKS cluster
- xip?

Read rest of the installation documents to:
- Install Kyma with Knative
- Reinstall Kyma
- Learn more about the installation scripts
