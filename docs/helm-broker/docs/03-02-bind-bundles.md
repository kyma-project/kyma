---
title: Bind bundles
type: Details
---

If you defined in the `meta.yaml` file that your plan is bindable, you must also create a `bind.yaml` file.
The `bind.yaml` file supports the [Service Catalog](https://github.com/kubernetes-incubator/service-catalog) binding concept. It is mandatory for all bindable plans as it contains information needed in the binding process. Currently, Kyma supports only the [credentials-type binding](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#types-of-binding).   

>**NOTE:** Resolving the values from the `bind.yaml` file is a post-provision action. If this operation ends with an error, the provisioning also fails.



### Template

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


### File specification

Set the following fields to create a valid `bind.yaml` file:

|   Field Name   |      Description                       |
|--------------|--------------------------------------------------------------|
| **credential** | A list of the credential variables returned during the binding action.  |
| **credential.name** | A name of the credential variable.  |
| **credential.value** | A variable value. You can also use the Helm Chart templating directives.  |
| **credential.valueFrom** | A source of the credential variable's value. You cannot use it if the value is not empty.  |
| **credential.valueFrom.configMapKeyRef** | A field which selects a ConfigMap key in the Helm chart release Namespace.    |
| **credential.valueFrom.configMapKeyRef.name**    | A name of the ConfigMap.  |
| **credential.valueFrom.configMapKeyRef.key**    | A name of the key from which the value is retrieved.  |
| **credential.valueFrom.secretKeyRef**  | A field which selects a Secret key in the Helm Chart release Namespace.     |
| **credential.valueFrom.secretKeyRef.name**    | A name of the Secret.           |
| **credential.valueFrom.secretKeyRef.key**    | A name of the key from which the value is retrieved. |
| **credential.valueFrom.serviceRef**   | A fields which selects a service resource in the Helm Chart release Namespace. |
| **credential.valueFrom.serviceRef.name**    | A name of the service.          |
| **credential.valueFrom.serviceRef.jsonpath**  | A JSONPath expression used to select the specified field value. For more information, see the [User Guide](https://kubernetes.io/docs/user-guide/jsonpath/). |
| **credentialFrom** | A list of sources to populate the credential variables on the binding action. When the key exists in multiple sources, the value associated with the last source takes precedence. Variables from the `credential` section override the values if duplicated keys exist. |
| **credentialFrom.configMapRef** | A ConfigMap to retrieve the values from. It must be available in the Helm chart release Namespace. |
| **credentialFrom.configMapRef.name**    | A name of the ConfigMap.   |
| **credentialFrom.secretRef** | A Secret to retrieve the values from. It must be available in the Helm chart release Namespace.  |
| **credentialFrom.secretRef.name**    | A name of the Secret.      |


See the example of the `bind.yaml` file:

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

### Credential name conflicts policy

The following rules apply in case of credential name conflicts:
- If the **credential** and **credentialFrom** fields have duplicate values, the system uses the values from the **credential** field.
- If you duplicate a key in the **credential** field, an error appears and informs you about the name of the key that the conflict refers to.
- If a key exists in the multiple sources defined by the **credentialFrom** section, the value associated with the last source takes precedence.
