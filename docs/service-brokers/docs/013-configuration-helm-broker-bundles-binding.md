---
title: Binding yBundles
type: Configuration
---


[bind]: https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#binding  "OSB Spec Binding"

If you defined in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file.
The `bind.yaml` file supports the [Service Catalog](https://github.com/kubernetes-incubator/service-catalog) binding concept. The `bind.yaml` file contains information the system uses in the [binding process][bind].
The `bind.yaml` file is mandatory for all bindable plans. Currently, Kyma supports only the [credentials](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#types-of-binding)-type binding.   


>**NOTE:** Resolving the values from the `bind.yaml` file is a post-provision action. If this operation ends with an error, the provisioning also fails.

## Details

This section provides an example of the `bind.yaml` file. It further describes the templating, the policy concerning credential name conflicts, and the detailed `bind.yaml` file specification.

### Example usage

```yaml
# bind.yaml
credential:
  - name: HOST
    value: redis.svc.cluster.local
  - name: PORT
    valueFrom:
      serviceRef:
        name: redis-svc
        jsonpath: '{ .spec.ports[?(@.name=="redis")].port }'
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: redis-secrets
        key: redis-password

credentialFrom:
  - configMapRef:
    name: redis-config
  - secretRef:
    name: redis-v2-secrets
```

In this example of the [binding action][bind], the Helm Broker returns the following values:
- A `HOST` key with the defined inlined value.
- A `PORT` key with the value from the field specified by the JSONPath expressions. The `redis-svc` Service runs this expression.
- A `REDIS_PASSWORD` key with a value selected by the `redis-password` key from the `redis-secrets` Secret.
- All the key-value pairs fetched from the `redis-config` ConfigMap.
- All the key-value pairs fetched from the `redis-v2-secrets` Secrets.

### Templating

In the `bind.yaml` file, you can use the Helm Chart templates directives.

```yaml
# bind.yaml
credential:
  - name: HOST
    value: {{ template "redis.fullname" . }}.{{ .Release.Namespace }}.svc.cluster.local
{{- if .Values.usePassword }}
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: {{ template "redis.fullname" . }}
        key: redis-password
{{- end }}
```
In this example, the system renders the `bind.yaml` file. The system resolves all the directives enclosed in the double curly braces in the same way as in the files located in the `templates` directory in your Helm chart.

### Credential name conflicts policy

The following rules apply in cases of credential name conflicts:
- When the `credential` and the `credentialFrom` sections have duplicate values, the system uses the values from the `credential` section.
- When you duplicate a key in the `credential` section, an error appears and informs you about the name of the key that the conflict refers to.
- When a key exists in the multiple sources defined by the `credentialFrom` section, the value associated with the last source takes precedence.

### File specification

|   Field Name   |                                                                                                                              Description                                                                                                                              |
|:--------------:|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|   [credential](#credential)   |                                                                                                         The list of the credential variables to return during the binding action.                                                                                                        |
| [credentialFrom](#credentialfrom) | The list of the sources to populate the credential variables on the binding action. When the key exists in multiple sources, the value associated with the last source takes precedence. The variables from the `credential` section override the values if duplicate keys exist. |

#### Credential

| Field Name |                                    Description                                    |
|:----------:|:---------------------------------------------------------------------------------:|
|    **name**    |                          The name of the credential variable.                         |
|    **value**   |      A variable value. You can also use the Helm Chart templating directives.      |
| [valueFrom](#valuefrom)  | The source of the credential variable's value. You cannot use it if the value is not empty. |

##### ValueFrom

|    Field Name   |                               Description                              |
|:---------------:|:----------------------------------------------------------------------:|
| [configMapKeyRef](#configmapkeyref) |    Selects a ConfigMap key in the Helm chart release Namespace.   |
|   [secretKeyRef](#secretkeyref)  |     Selects a Secret key in the Helm Chart release Namespace.     |
|    [serviceRef](#serviceref)   | Selects a Service resource in the Helm Chart release Namespace. |

###### ConfigMapKeyRef

| Field Name |                            Description                            |
|:----------:|:-----------------------------------------------------------------:|
|    **name**    |                      The name of the ConfigMap.                       |
|     **key**    |   The name of the key from which the value is retrieved.  |

###### SecretKeyRef

| Field Name |                            Description                            |
|:----------:|:-----------------------------------------------------------------:|
|    **name**    |                       The name of the Secret.                          |
|     **key**    | The name of the key from which the value is retrieved. |

###### ServiceRef

| Field Name |                                                                    Description                                                                    |
|:----------:|:-------------------------------------------------------------------------------------------------------------------------------------------------:|
|    **name**    |                                                                The name of the Service.                                                               |
|  **jsonpath**  | The JSONPath expression used to select the specified field value. For more information, see the [User Guide](https://kubernetes.io/docs/user-guide/jsonpath/). |

#### CredentialFrom

|  Field Name  |                                   Description                                 |
|:------------:|:-----------------------------------------------------------------------------:|
| [configMapRef](#configmapref) | The ConfigMap to retrieve the values from. It must be available in the Helm Chart release Namespace. |
|   [secretRef](#secretref)  |   The Secret to retrieve the values from. It must be available in the Helm Chart release Namespace.  |

##### ConfigMapRef

| Field Name |         Description        |
|:----------:|:--------------------------:|
|    **name**    |    The name of the ConfigMap.   |

##### SecretRef

| Field Name |        Description        |
|:----------:|:-------------------------:|
|    **name**    |  The name of the Secret.      |
