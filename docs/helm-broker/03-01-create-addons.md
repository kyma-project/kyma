---
title: Create addons
type: Details
---

Addons which the Helm Broker uses must have a specific structure. These are all possible files that you can include in your addons:

```
sample-addon/
   ├── meta.yaml                             # [REQUIRED] A file which contains metadata information about this addon
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
   └── docs/                                 # A directory which contains documentation for this addon
        ├── meta.yaml                        # [REQUIRED] A file which contains metadata information about documentation for this addon
        ├── {assets}                         # Files with documentation and assets
        └── ....
```

> **NOTE:** All file names in an addon repository are case-sensitive.

For details about particular files, read the following sections.

## meta.yaml file

The `meta.yaml` file contains information about the addon. Define the following fields to create a service object which complies with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#service-object).

|      Field Name     | Required |                   Description             |
|-------------------|:--------:|----------------------------------------------|
|         **name**        |   Yes   | The name of the addon.  |
|       **version**       |   Yes   | The version of the addon. It is a broker service identifier.  |
|          **id**         |   Yes   | The broker service identifier.  |
|     **description**     |   Yes   | The short description of the service. |
|     **displayName**     |   Yes   | The display name of the addon.    |
|         **tags**        |   No  | Keywords describing the provided service, separated by commas.     |
|       **bindable**      |   No  | The field that specifies whether you can bind a given addon. |
| **providerDisplayName** |   No  | The name of the upstream entity providing the actual service.  |
|   **longDescription**   |   No  | The long description of the service.     |
|   **documentationURL**  |   No  | The link to the documentation page for the service.        |
|      **supportURL**     |   No  | The link to the support page for the service.     |
|       **imageURL**      |   No  | The URL to an image. You must provide the image in the `SVG` format.          |
|       **labels**        |   No  | Key-value pairs that help you to organize your project. Use labels to indicate different elements, such as Namespaces, services, or teams.   |
| **bindingsRetrievable** |   No  | The field that specifies whether fetching a ServiceBinding using a GET request on the resource's endpoint is supported for all plans. The default value is `false`.   |
|   **planUpdatable**     |   No  |  The field that specifies whether instances of this service can be updated to a different plan. The default value is `false`  |
|       **requires**      |   No  | The list of permissions the user must grant to the instances of this service. |
| **provisionOnlyOnce**   |   No  | The field that specifies whether the addon can be provisioned only once in a given Namespace. The default value is `false`. |

> **NOTE**: The **provisionOnlyOnce** and **local** keys are reserved and cannot be added to the **labels** entry, since the Helm Broker overrides them at runtime. The Helm Broker always adds the `local:true` label and it adds the `provisionOnlyOnce:true` label only if **provisionOnlyOnce** is set to `true`.

## chart directory

In the `chart` directory, create a folder with the same name as your chart. Put all the files related to your chart in this folder. The system supports Helm version 2.6.

> **NOTE:** The Helm Broker uses the [helm wait](https://github.com/kubernetes/helm/blob/release-2.6/docs/using_helm.md#helpful-options-for-installupgraderollback) option to ensure that all the resources that a chart creates are available. If you set your Deployment **replicas** to `1`, you must set **maxUnavailable** to `0` as a part of the rolling update strategy.

## plans directory

The `plans` directory must contain at least one plan. Each plan must contain the `meta.yaml` file. Other files are not mandatory.

* `meta.yaml` file - contains information about a given plan. Define the following fields to create a plan object which complies with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md#plan-object).

|  Field Name | Required |      Description               |
|-----------|:--------:|------------------------------------|
|     **name**    |   Yes   |     The name of the plan.   |
|      **id**     |   Yes   |     The ID of the plan. |
| **description** |   Yes   | The description of the plan. |
| **displayName** |   Yes   | The display name of the plan. |
|  **bindable**   |   No  | The field that specifies whether you can bind an instance of the plan or not. The default value is `false`. |
|     **free**    |   No  | The attribute which specifies whether an instance of the plan is free or not. The default value is `false`.    |

* `bind.yaml` file - contains information about binding in a specific plan. If you define in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file. For more information about this file, see [this](#details-bind-addons) document.

* `values.yaml` file - provides the default configuration values in a given plan for the chart definition located in the `chart` directory. For more information, see the [values files](https://github.com/kubernetes/helm/blob/release-2.6/docs/chart_template_guide/values_files.md) specification.

* `create-instance-schema.json` file - contains a schema that defines parameters for a provision operation of a ServiceInstance. Each input parameter is expressed as a property within a JSON object.

* `update-instance-schema.json` file - contains a schema that defines parameters for an update operation of a ServiceInstance. Each input parameter is expressed as a property within a JSON object.

* `bind-instance-schema.json` file - contains a schema that defines parameters for a bind operation. Each input parameter is expressed as a property within a JSON object.

>**NOTE:** For more information about schemas, see [this](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#schemas-object) specification.

## docs directory

In the `docs` directory, provide documentation for your addon. The documentation can include Markdown documents, AsyncAPI, OData, and OpenAPI specification files. Create the `assets` directory inside the `docs` directory to store assets, such as images. The `docs` directory must contain a `meta.yaml` file, which provides information on how documentation for the addon is uploaded.
Because you can install the Helm Broker as a ClusterServiceBroker or as a ServiceBroker, documentation for addons is provided using either [ClusterDocsTopics](/components/headless-cms/#custom-resource-clusterdocstopic) or [DocsTopics](/components/headless-cms/#custom-resource-docs-topic) custom resources, respectively.

The `meta.yaml` file contains the specification of the ClusterDocsTopic or DocsTopic. The example structure of the `meta.yaml` file looks as follows:

|  Field Name | Required |      Description               |
|-----------|:--------:|------------------------------------|
|   **docs[]**                           |   Yes   | Contains the definitions of documentation.   |
| **docs[].template**                    |   Yes   | Contains the specification of the ClusterDocsTopic or DocsTopic. |
| **docs[].template.displayName**        |   Yes   | Specifies the display name of the ClusterDocsTopic or DocsTopic. |
| **docs[].template.description**        |   Yes   | Provides the description of the ClusterDocsTopic or DocsTopic. |
| **docs[].template.sources[]**          |   Yes   | Contains the definitions of assets for an addon. |
| **docs[].template.sources[].type**     |   Yes   | Defines the type of the asset. |
| **docs[].template.sources[].name**     |   Yes   | Defines a unique identifier of a given asset. It must be unique for a given asset type. |
| **docs[].template.sources[].mode**     |   Yes   | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR format. Use `single` for one file and `package` for a set of files. |
| **docs[].template.sources[].url**      |   Yes   | Specifies the location of a file. |
| **docs[].template.sources[].filter**   |   Yes   | Specifies the directory from which the documentation is fetched. The regex used in the filter must be [RE2](https://golang.org/s/re2syntax)-compliant.  |

>**NOTE:** Currently you can provide only one entry in the `docs` array.

See [this](https://github.com/kyma-project/addons/tree/master/addons/testing-0.0.1/docs) example of the `docs` directory with documentation for the testing addon. For more information on how to provide addons documentation, read [this](#details-provide-service-classes-documentation) document.

## Troubleshooting

Use the dry run mode to check the generated manifests of the chart without installing it.
The `--debug` option prints the generated manifests.
As a prerequisite, you must install [Helm](https://github.com/kubernetes/helm) on your machine to run this command:

```
 helm install --dry-run {path-to-chart} --debug
```
For more details, read the Helm [official documentation](https://v2.helm.sh/docs/chart_template_guide/#debugging-templates).
