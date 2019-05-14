---
title: Create a bundle
type: Details
---

Bundles which the Helm Broker uses must have a specific structure. These are all possible files that you can include in your bundle:

```
sample-bundle/
   ├── meta.yaml                             # [REQUIRED] A file which contains metadata information about this bundle
   ├── chart/                                # [REQUIRED] A directory which contains a Helm chart that installs your Kubernetes resources
   │    └── {chart-name}/                    # [REQUIRED] A Helm chart directory
   │         └── ....                        # [REQUIRED] Helm chart files   
   ├── plans/                                # [REQUIRED] A directory which contains the possible plans for an installed chart
   │    ├── example-enterprise               # [REQUIRED] A directory which contains files for a specific plan
   │    │   ├── meta.yaml                    # [REQUIRED] A file which contains metadata information about this plan
   │    │   ├── bind.yaml                    # A file which contains information required to bind this plan
   │    │   ├── create-instance-schema.json  # JSON schema definitions for creating a ServiceInstance
   │    │   ├── bind-instance-schema.json    # JSON schema definitions for binding a ServiceInstance
   │    │   ├── update-instance-schema.json  # JSON schema definitions for updating a ServiceInstance
   │    │   └── values.yaml                  # Default configuration values in this plan for a chart defined in the `chart` directory
   │    └── ....
   │
   └── docs/                                 # A directory which contains documentation for this bundle
        ├── meta.yaml                        # [REQUIRED] A file which contains metadata information about documentation for this bundle
        ├── {assets}                         # Files with documentation and assets
        └── ....
```

> **NOTE:** All file names in a bundle repository are case-sensitive.

For details about particular files, read the following sections.

## meta.yaml file

The `meta.yaml` file contains information about the bundle. Define the following fields to create a service object which complies with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#service-object).

|      Field Name     | Required |                   Description             |
|-------------------|:--------:|----------------------------------------------|
|         **name**        |   YES   | The name of the bundle.  |
|       **version**       |   YES   | The version of the bundle. It is a broker service identifier.  |
|          **id**         |   YES   | The broker service identifier.  |
|     **description**     |   YES   | The short description of the service. |
|     **displayName**     |   YES   | The display name of the bundle.    |
|         **tags**        |   NO  | Keywords describing the provided service, separated by commas.     |
|       **bindable**      |   NO  | The field that specifies whether you can bind a given bundle. |
| **providerDisplayName** |   NO  | The name of the upstream entity providing the actual service.  |
|   **longDescription**   |   NO  | The long description of the service.     |
|   **documentationURL**  |   NO  | The link to the documentation page for the service.        |
|      **supportURL**     |   NO  | The link to the support page for the service.     |
|       **imageURL**      |   NO  | The URL to an image. You must provide the image in the `SVG` format.          |
|       **labels**        |   NO  | Key-value pairs that help you to organize your project. Use labels to indicate different elements, such as Namespaces, services, or teams.   |
| **bindingsRetrievable** |   NO  | The field that specifies whether fetching a ServiceBinding using a GET request on the resource's endpoint is supported for all plans. The default value is `false`.   |
|   **planUpdatable**     |   NO  |  The field that specifies whether instances of this service can be updated to a different plan. The default value is `false`  |
|       **requires**      |   NO  | The list of permissions the user must grant to the instances of this service. |
| **provisionOnlyOnce**   |   NO  | The field that specifies whether the bundle can be provisioned only once in a given Namespace. The default value is `false`. |

> **NOTE**: The **provisionOnlyOnce** and **local** keys are reserved and cannot be added to the **labels** entry, since the Helm Broker overrides them at runtime. The Helm Broker always adds the `local:true` label and it adds the `provisionOnlyOnce:true` label only if **provisionOnlyOnce** is set to `true`.

## chart directory

In the `chart` directory, create a folder with the same name as your chart. Put all the files related to your chart in this folder. The system supports Helm version 2.6.

> **NOTE:** The Helm Broker uses the [helm wait](https://github.com/kubernetes/helm/blob/release-2.6/docs/using_helm.md#helpful-options-for-installupgraderollback) option to ensure that all the resources that a chart creates are available. If you set your Deployment **replicas** to `1`, you must set **maxUnavailable** to `0` as a part of the rolling update strategy.

## plans directory

The `plans` directory must contain at least one plan. Each plan must contain the `meta.yaml` file. Other files are not mandatory.

* `meta.yaml` file - contains information about a given plan. Define the following fields to create a plan object which complies with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#plan-object).

|  Field Name | Required |      Description               |
|-----------|:--------:|------------------------------------|
|     **name**    |   YES   |     The name of the plan.   |
|      **id**     |   YES   |     The ID of the plan. |
| **description** |   YES   | The description of the plan. |
| **displayName** |   YES   | The display name of the plan. |
|  **bindable**   |   NO  | The field that specifies whether you can bind an instance of the plan or not. The default value is `false`. |
|     **free**    |   NO  | The attribute which specifies whether an instance of the plan is free or not. The default value is `false`.    |

* `bind.yaml` file - contains information about binding in a specific plan. If you define in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file. For more information about this file, see [this](#details-bind-bundles) document.

* `values.yaml` file - provides the default configuration values in a given plan for the chart definition located in the `chart` directory. For more information, see the [values files](https://github.com/kubernetes/helm/blob/release-2.6/docs/chart_template_guide/values_files.md) specification.

* `create-instance-schema.json` file - contains a schema that defines parameters for a provision operation of a ServiceInstance. Each input parameter is expressed as a property within a JSON object.

* `update-instance-schema.json` file - contains a schema that defines parameters for an update operation of a ServiceInstance. Each input parameter is expressed as a property within a JSON object.

* `bind-instance-schema.json` file - contains a schema that defines parameters for a bind operation. Each input parameter is expressed as a property within a JSON object.

>**NOTE:** For more information about schemas, see [this](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#schemas-object) specification.

## docs directory

In the `docs` directory, provide documentation for your bundle. The `docs` directory must contain a `meta.yaml` which provides information on how documentation for the bundle is uploaded.
As the Helm Broker is installed as a ClusterServiceBroker, documentation for bundles is provided using [ClusterDocsTopics](/components/headless-cms/#custom-resource-clusterdocstopic).

The `meta.yaml` file contains the specification of the ClusterDocsTopic. The example structure of the `meta.yaml` file looks as follows:

|  Field Name | Required |      Description               |
|-----------|:--------:|------------------------------------|
|   **docs[]**                           |   YES   | Contains the definitions of documentation.   |
| **docs[].template**                    |   YES   | Contains the specification of the ClusterDocsTopic. |
| **docs[].template.displayName**        |   YES   | Specifies the display name of the ClusterDocsTopic. |
| **docs[].template.description**        |   YES   | Provides the description of the ClusterDocsTopic. |
| **docs[].template.sources[]**          |   YES   | Contains the definitions of assets for a bundle. |
| **docs[].template.sources[].type**     |   YES   | Defines the type of the asset. |
| **docs[].template.sources[].name**     |   YES   | Defines the name of the asset. |
| **docs[].template.sources[].mode**     |   YES   | Defines the type of the asset file. |
| **docs[].template.sources[].url**      |   YES   | Defines the URL under which the asset is stored. |
| **docs[].template.sources[].filter**   |   YES   | Defines the path from which to upload assets.  |

>**NOTE:** Currently you can provide only one entry in the `docs` array.

Using the Helm Broker, you can provision a broker which provides its own Service Classes. To learn how to upload documentation for those classes, read [this](#details-bundles-docs.md) document.

## Troubleshooting

Use the dry run mode to check the generated manifests of the chart without installing it.
The `--debug` option prints the generated manifests.
As a prerequisite, you must install [Helm](https://github.com/kubernetes/helm) on your machine to run this command:

```
 helm install --dry-run {path-to-chart} --debug
```
For more details, read the Helm [official documentation](https://docs.helm.sh/chart_template_guide/#debugging-templates).
