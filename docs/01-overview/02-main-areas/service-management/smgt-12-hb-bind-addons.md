---
title: Bind addons
---

If you defined in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file.
The `bind.yaml` file supports the Service Catalog binding concept. It is mandatory for all bindable plans as it contains information needed during the binding process. Currently, Kyma supports only the [credentials-type binding](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#types-of-binding).   

>**NOTE:** Resolving the values from the `bind.yaml` file is a post-provision action. If this operation ends with an error, the provisioning also fails.

In the `bind.yaml` file, you can use the Helm chart templates directives. See the example:

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


## File specification

Define the following fields to create a valid `bind.yaml` file:

|   Field Name   |      Description                       |
|--------------|--------------------------------------------------------------|
| **credential** | The list of credential variables returned during the binding action.  |
| **credential.name** | The name of a given credential variable.  |
| **credential.value** | The variable value. You can also use the Helm Chart templating directives. This field is interchangeable with **credential.valueFrom**. |
| **credential.valueFrom** | The source of the given credential variable's value. This field is interchangeable with **credential.value**.  |
| **credential.valueFrom.configMapKeyRef** | The field which selects a ConfigMap key in the Helm chart release Namespace.    |
| **credential.valueFrom.configMapKeyRef.name** | The name of the ConfigMap.  |
| **credential.valueFrom.configMapKeyRef.key**  | The name of the key from which the value is retrieved.  |
| **credential.valueFrom.secretKeyRef**  | The field which selects a Secret key in the Helm Chart release Namespace.     |
| **credential.valueFrom.secretKeyRef.name**    | The name of the Secret.     |
| **credential.valueFrom.secretKeyRef.key**    | The name of the key from which the value is retrieved. |
| **credential.valueFrom.serviceRef**   | The field which selects a service resource in the Helm Chart release Namespace. |
| **credential.valueFrom.serviceRef.name**    | The name of the service.          |
| **credential.valueFrom.serviceRef.jsonpath**  | The JSONPath expression used to select the specified field value. For more information, see the [User Guide](https://kubernetes.io/docs/user-guide/jsonpath/). |
| **credentialFrom** | The list of sources to populate credential variables on the binding action. When the key exists in multiple sources, the value associated with the last source takes precedence. Variables from the `credential` section override the values if duplicated keys exist. |
| **credentialFrom.configMapRef** | The ConfigMap to retrieve the values from. It must be available in the Helm chart release Namespace. |
| **credentialFrom.configMapRef.name**    | The name of the ConfigMap.   |
| **credentialFrom.secretRef** | The Secret to retrieve the values from. It must be available in the Helm chart release Namespace.  |
| **credentialFrom.secretRef.name**    | The name of the Secret.      |


See the fully extended example of the `bind.yaml` file:

```yaml
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
  - name: REDIS_DB_NAME
    valueFrom:
      configMapKeyRef:
        name: redis-cm
        key: redis-db-name

credentialFrom:
  - configMapRef:
    name: redis-config
  - secretRef:
    name: redis-v2-secrets
```

In this example, the Helm Broker returns the following values:
- A `HOST` key with the defined inlined value.
- A `PORT` key with the value from the field specified by the JSONPath expressions. The `redis-svc` Service runs this expression.
- A `REDIS_PASSWORD` key with a value selected by the `redis-password` key from the `redis-secrets` Secret.
- All the key-value pairs fetched from the `redis-config` ConfigMap.
- All the key-value pairs fetched from the `redis-v2-secrets` Secrets.


## Credential name conflicts policy

The following rules apply in case of credential name conflicts:
- If the **credential** and **credentialFrom** fields have duplicate values, the system uses the values from the **credential** field.
- If you duplicate a key in the **credential** field, an error appears and informs you about the name of the key that the conflict refers to.
- If a key exists in the multiple sources defined by the **credentialFrom** section, the value associated with the last source takes precedence.
