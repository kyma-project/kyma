---
title: Create a bundle
type: Configuration
---

[bind]: https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#binding  "OSB Spec Binding"
[service-objects]: https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#service-object "OSB Spec Service Objects"
[service-metadata]: https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata "OSB Spec Service Metadata"
[plan-objects]: https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#plan-object "OSB Spec Plan Objects"

To create your own bundle, you must create a directory with the following structure:

```
sample-bundle/
  ├── meta.yaml                             # A file which contains the metadata information about this bundle
  ├── chart/                                # A directory which contains a Helm chart that installs your Kubernetes resources
  │    └── <chart-name>/                    # A Helm chart directory
  │         └── ....                        # Helm chart files
  └── plans/                                # A directory which contains the possible plans for an installed chart
       ├── example-enterprise               # A directory of files for a specific plan
       │   ├── meta.yaml                    # A file which contains the metadata information about this plan
       │   ├── bind.yaml                    # A file which contains information about the values that the Helm Broker returns when it receives the bind request
       │   ├── create-instance-schema.json  # The JSON Schema definitions for creating a ServiceInstance
       │   ├── bind-instance-schema.json    # The JSON Schema definitions for binding a ServiceInstance
       │   ├── update-instance-schema.json  # The JSON Schema definitions for updating a ServiceInstance
       │   └── values.yaml                  # The default configuration values in this plan for a chart defined in the `chart` directory
       └── ....
```

> **NOTE:** All the file names in the bundle directory are case-sensitive.


### The bundle meta.yaml file

The `meta.yaml` file is mandatory as it contains information about the bundle. Set the following fields to create service objects which comply with the [Open Service Broker API][service-objects].

|      Field Name     | Required |                                                                  Description                                                                           |
|:-------------------:|:--------:|:------------------------------------------------------------------------------------------------------------------------------------------------------:|
|         **name**        |   true   |                           The bundle name. It has the same restrictions as defined in the [Open Service Broker API][service-objects].                           |
|       **version**       |   true   | The bundle version. It is a broker service identifier. It has the same restrictions as defined in the [Open Service Broker API][service-objects]. |
|          **id**         |   true   |            A broker service identifier. It has the same restrictions as defined in the [Open Service Broker API][service-objects].           |
|     **description**     |   true   |                  A short description of the service. It has the same restrictions as defined in the [Open Service Broker API][service-objects].                  |
|         **tags**        |   false  |                                    The keywords describing the provided service, separated by commas.                                                          |
|       **bindable**      |   false  |                                    The bindable field described in the [Open Service Broker API][service-metadata].                                          |
|     **displayName**     |   true   |                                    The **displayName** field described in the [Open Service Broker API][service-metadata].                                       |
| **providerDisplayName** |   false  |                                The **providerDisplayName** field described in the [Open Service Broker API][service-metadata].                                   |
|   **longDescription**   |   false  |                                  The **longDescription** field described in the [Open Service Broker API][service-metadata].                                     |
|   **documentationURL**  |   false  |                                  The **documentationURL** field described in the [Open Service Broker API][service-metadata].                                    |
|      **supportURL**     |   false  |                                     The **supportURL** field described in the [Open Service Broker API][service-metadata].                                       |
|       **imageURL**      |   false  |     The **imageURL** field described in the [Open Service Broker API][service-metadata]. You must provide the image as an SVG.          |
|       **labels**        |   false  |     To organize your project, add arbitrary labels as key/value pairs. Use labels to indicate different elements, such as Namespaces, services, or teams.   |
| **bindingsRetrievable** |   false  |     The **bindingRetrievable** field specifies whether fetching a ServiceBinding using a `GET` call on the resource's endpoint is supported for all plans. The default value is `false`.   |
|   **planUpdatable**     |   false  |     The **planUpdatable** field specifies whether instances of this service can be updated to a different plan. The default value is `false`  |
|       **requires**      |   false  |     The **requires** field defines a list of permissions the user must grant to the instances of this service.          |
| **provisionOnlyOnce**   |   false  |     Defines that the bundle can be provisioned only once in a given Namespace. The default value is `false` |

> **NOTE**: The **provisionOnlyOnce** and **local** keys are reserved and cannot be added to the **labels** entry, since the Helm Broker overrides them at runtime. The Helm Broker always adds the `local:true` label and it adds the `provisionOnlyOnce:true` label only if **provisionOnlyOnce** is set to `true`.

### The chart directory

In the mandatory `chart` directory, create a folder with the same name as your chart. Put all the files related to your chart in this folder. The system supports chart version 2.6.

If you are not familiar with the chart definitions, see the [Charts](https://github.com/kubernetes/helm/blob/release-2.6/docs/charts.md) specification.

> **NOTE:** Helm Broker uses the [helm wait](https://github.com/kubernetes/helm/blob/release-2.6/docs/using_helm.md#helpful-options-for-installupgraderollback) option to ensure that all the resources that a chart creates are available. If you set your Deployment **replicas** to `1`, you must set **maxUnavailable** to `0` as a part of the rolling update strategy.

### The plans directory

The mandatory `plans` directory must contain at least one plan.
A directory for a specific plan must contain the `meta.yaml` file. Other files, such as `create-instance-schema.json`, `bind.yaml` and `values.yaml` are not mandatory.

#### The meta.yaml file

The `meta.yaml` file contains information about a bundle plan. Set the following fields to create the plan objects, which comply with the [Open Service Broker API][plan-objects].

|  Field Name | Required |                                             Description                                                    |
|:-----------:|:--------:|:----------------------------------------------------------------------------------------------------------:|
|     **name**    |   true   |     The plan name. It has the same restrictions as defined in the [Open Service Broker API][plan-objects].    |
|      **id**     |   true   |      The plan ID. It has the same restrictions as defined in the [Open Service Broker API][plan-objects].     |
| **description** |   true   | The plan description. It has the same restrictions as defined in the [Open Service Broker API][plan-objects]. |
| **displayName** |   true   | The plan display name. It has the same restrictions as defined in the [Open Service Broker API][plan-objects]. |
|  **bindable**   |   false  | The plan bindable attribute. It has the same restrictions as defined in the [Open Service Broker API][plan-objects].    |
|     **free**    |   false  | The attribute which specifies whether the instance is free or not. The default value is `false`.    |

#### The bind.yaml file

The `bind.yaml` file contains the information required for the [binding action][bind] in a specific plan.
If you defined in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file.
For more information about the content of the `bind.yaml` file, see the [Binding bundles](#configuration-binding-bundles) document.

#### The values.yaml file

The `values.yaml` file provides the default configuration values in a concrete plan for the chart definition located in the `chart` directory.
This file is not required.
For more information about the content of the `values.yaml` file, see the [Values Files](https://github.com/kubernetes/helm/blob/release-2.6/docs/chart_template_guide/values_files.md) specification.

#### The create-instance-schema.json file

The `create-instance-schema.json` file contains a schema used to define the parameters for a provision operation of the ServiceInstance. Each input parameter is expressed as a property within a JSON object.
This file is not required.
For more information about the content of the `create-instance-schema.json` file, see the [Schemas](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#schemas-object) specification.

#### The update-instance-schema.json file

The `update-instance-schema.json` file contains a schema used to define the parameters for an update operation of the ServiceInstance. Each input parameter is expressed as a property within a JSON object.
This file is not required.
For more information about the content of the `update-instance-schema.json` file, see the [Schemas](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#schemas-object) specification.

#### The bind-instance-schema.json file

The `bind-instance-schema.json` file contains a schema used to define the parameters for a bind operation. Each input parameter is expressed as a property within a JSON object.
This file is not required.
For more information about the content of the `bind-instance-schema.json` file, see the [Schemas](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#schemas-object) specification.

### Troubleshooting

Use the dry-run mode to check the generated manifests of the chart without installing it.
The **--debug** option prints the generated manifests.
As a prerequisite, you must install [Helm](https://github.com/kubernetes/helm) on your machine to run this command:

```
 helm install --dry-run {path-to-chart} --debug
```
For more details, read the Helm [official documentation](https://docs.helm.sh/chart_template_guide/#debugging-templates).
