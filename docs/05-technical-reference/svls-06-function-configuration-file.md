---
title: Function configuration file
---

When you initialize a Function (with the `init` command), CLI creates the `config.yaml` file in your workspace folder. This file contains the whole Function's configuration and specification not only for the Function custom resource but also any other related resources you create for it, such as Subscriptions and APIRules.

## Specification for an inline Function

See the sample `config.yaml` for an inline Function for which code and dependencies are stored in the Function custom resource (CR) under the **spec.source** and **spec.deps** fields. This specification also contains the definition of a sample Subscription and APIRules for the Function:

```yaml
name: function-practical-filip5
namespace: testme
runtime: nodejs14
labels:
    app: serverless-test
source:
    sourceType: inline
    sourcePath: /tmp/cli
    sourceHandlerName: /code/handler.js
    depsHandlerName: /dependencies/package.json
resources:
    limits:
      cpu: 1
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 500Mi
subscriptions:
    - name: function-practical-filip5
      protocol: ""
      filter:
        dialect: ""
        filters:
            - eventSource:
                property: source
                type: exact
                value: ""
              eventType:
                property: type
                type: exact
                value: sap.kyma.custom.demo-app.order.created.v1
apiRules:
    - name: function-practical-filip5
      gateway: kyma-gateway.kyma-system.svc.cluster.local
      service:
        host: path.34.90.136.181.xip.io
        port: 80
      rules:
        - methods:
            - GET
            - POST
            - PUT
            - PATCH
            - DELETE
            - HEAD
          accessStrategies: []
        - path: /path1/something1
          methods:
            - PUT
            - PATCH
            - DELETE
          accessStrategies:
            - handler: noop
        - path: /path1/something2
          methods:
            - GET
          accessStrategies:
            - config:
                required_scope: ["read"]
              handler: oauth2_introspection
        - path: /path2
          methods:
            - DELETE
          accessStrategies:
              handler: jwt
            - config:
                jwksUrls:
                    - http://dex-service.kyma-system.svc.cluster.local:5556/keys
                    - http://dex-service.kyma-system.svc.cluster.local:5556
                trustedIssuers:
                    - https://dex.34.90.136.181.xip.io
                    - https://dex.34.90.136.181.xip.io
env:
    - name: REDIS_PASS
      value: YgJUg8z6eA
    - name: REDIS_PORT
      value: "6379"
    - name: REDIS_HOST
      value: hb-redis-enterp-6541066a-edbc-422f-8bef-fafca0befea8-redis.testme.svc.cluster.local
    - valueFrom:
        configMapKeyRef:
          Name: configmap1
          Key: token-field
    - valueFrom:
        secretKeyRef:
          Name: secret1
          Key: token-field
```

## Specification for a Git Function

See the sample `config.yaml` for a [Git Function](./05-technical-reference/svls-04-git-source-type.md) for which code and dependencies are stored in a selected Git repository:

```yaml
name: function-practical-marcin
namespace: iteration-review
runtime: nodejs14
source:
    sourceType: git
    url: https://github.com/username/public-gitops.git
    repository: my-repo
    reference: main
    baseDir: /
    credentialsSecretName: secret2
```

## Parameters

See all parameter descriptions.

>**NOTE:** The **Default value** column specifies the values that Kyma CLI sets when applying resources on a cluster, if no other values are provided.

| Parameter         | Required | Related custom resource | Default value  | Description|
| ---------------------------------------- | :------------: | ---------| ---------| ------- |
| **name**              | Yes | Function | | Specifies the name of your Function.         |
| **namespace**             | No | Function | `default` | Defines the Namespace in which the Function is created.        |
| **runtime**             | Yes | Function | | Specifies the execution environment for your Function. The available values are `nodejs12`, `nodejs14`, and `python38`.      |
| **labels**             | No | Function | | Specifies the Function's Pod labels. |
| **source**            | Yes | Function | | Provides details on the type and location of your Function's source code and dependencies.         |
| **source.sourceType**            | Yes | Function | | Defines whether you use either inline code or a Git repository as the source of the Function's code and dependencies. It must be set either to `inline` or `git`.         |
| **source.sourcePath**             | No | Function | Location of the `config.yaml` file | Specifies the absolute path to the directory with the Function's source code.         |
| **source.sourceHandlerName**             | No | Function | `handler.js` (Node.js) or `handler.py` (Python) | Defines the path to the file with your Function's code. Specify it if you want to store source code separately from the `config.yaml`.  This path is a relative path to the one provided in **source.sourcePath**. |
| **source.depsHandlerName**             | No | Function | `package.json` (Node.js) or `requirements.txt` (Python) | Defines the path to the file with your Function's dependencies. Specify it if you want to store dependencies separately from the `config.yaml`. This path is a relative path to the one provided in **source.sourcePath**. |
| **source.url**             | No | Function | | Provides the address to the Git repository with the Function's code and dependencies. Depending on whether the repository is public or private and what authentication method is used to access it, the URL must start with the `http(s)`, `git`, or `ssh` prefix, and end with the `.git` suffix.  |
| **source.repository**             | No | Function | Function name | Specifies the name of the Git repository.  |
| **source.reference**             | No | Function | | Specifies either the branch name or the commit revision from which the Function Controller automatically fetches the changes in the Function's code and dependencies.  |
| **source.baseDir**             | No | Function | | Specifies the location of your code dependencies in the repository. It is recommended to keep the source files at the root of your repository (`/`).  |
| **source.credentialsSecretName**             | No | Function | | Specifies the name of the Secret with credentials to the Git repository. It is used by the Function Controller to authenticate to the Git repository to fetch the Function's source code and dependencies. This Secret must be stored in the same Namespace as the [GitRepository CR](./05-technical-reference/06-custom-resources/svls-02-gitrepository.md). |
| **resources**             | No | Function | | Defines CPU and memory available for the Function's Pod to use.  |
| **resources.limits**             | No | Function | | Defines the maximum available CPU and memory values for the Function.  |
| **resources.limits.cpu**             | No | Function | `100m` | Defines the maximum available CPU value for the Function.  |
| **resources.limits.memory**             | No | Function | `128Mi` | Defines the maximum available memory value for the Function.  |
| **resources.requests**              | No | Function | | Defines the minimum requested CPU and memory values for a Function.  |
| **resources.requests.cpu**              | No | Function | `50m` | Defines the minimum requested CPU value for the Function.  |
| **resources.requests.memory**              | No | Function | `64Mi` | Defines the minimum requested memory value for the Function.  |
| **subscriptions**   | No | Subscription | | Defines a Subscription by which the Function gets triggered to perform a business logic defined in the Function's source code.  |
| **subscriptions.name**           |  Yes | Subscription | Function name | Specifies the name of the Subscription custom resource. It takes the name from the Function unless you specify otherwise.    |
| **subscriptions.protocol**           | Yes  | Subscription | | Defines the rules and formats applied for exchanging messages between the components of a given messaging system. Subscriptions in Kyma CLI use the [NATS](https://docs.nats.io/) messaging protocol by default. Must be set to `""`.         |
| **subscriptions.filter** | No | Subscription | | Defines the filter for the Subscriptions. |
| **subscriptions.filter.dialect**            | No | Subscription | | Indicates the filter expression language supported by an event producer. Subscriptions specifying the **filter** property must specify the dialect as well. All other properties are dependent on the dialect being used. In the current implementation, this field is treated as a constant which is blank.    |
| **subscriptions.filter.filters**            | Yes | Subscription | | Specifies the filtering parameters for the given event.   |
| **subscriptions.filter.filters.eventSource**            | Yes | Subscription | | Defines the origin from which events are published.   |
| **subscriptions.filter.filters.eventSource.property**            | Yes | Subscription | | Must be set to `source`.   |
| **subscriptions.filter.filters.eventSource.type**           | Yes  | Subscription | | Must be set to `exact`.    |
| **subscriptions.filter.filters.eventSource.value**            | Yes | Subscription | | Must be set to `""` for the NATS backend.  |
| **subscriptions.filter.filters.eventType**           | Yes  | Subscription | | Defines the type of events used to trigger workloads.     |
| **subscriptions.filter.filters.eventType.property**           | Yes  | Subscription | | Must be set to `type`.    |
| **subscriptions.filter.filters.eventType.type**          |  No  | Subscription | | Must be set to `exact`.    |
| **subscriptions.filter.filters.eventType.value**           | Yes  | Subscription | | Defines the name of the event the Function is subscribed to, for example `sap.kyma.custom.demo-app.order.created.v1`.         |
| **apiRules**    | No | APIRule | | Provides the rules defining how your Function's Service API can be accessed. |
| **apiRules.name**            | Yes | APIRule | Function name | Specifies the name of the exposed Service. It takes the name from the Function unless you specify otherwise.        |
| **apiRules.gateway**            | No | APIRule | `kyma-gateway.kyma-system.svc.cluster.local` | Specifies the [Istio Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/).   |
| **apiRules.service**            | No | APIRule | | Specifies the name of the exposed Service.     |
| **apiRules.service.host**            | No | APIRule | | Specifies the Service's communication address for inbound external traffic.         |
| **apiRules.service.port**            | No | APIRule | `80`. | Defines the port on which the Function's Service is exposed. This value cannot be modified. |
| **apiRules.rules**            | Yes | APIRule | | Specifies the array of [Oathkeeper](https://www.ory.sh/oathkeeper/) access rules.         |
| **apiRules.rules.methods**            | No | APIRule | | Specifies the list of HTTP request methods available for **apiRules.rules.path** .        |
| **apiRules.rules.accessStrategies**            | Yes | APIRule | | Specifies the array of [Oathkeeper authenticators](https://www.ory.sh/oathkeeper/docs/pipeline/authn/). The supported authenticators are `oauth2_introspection`, `jwt`, `noop`, and `allow`.         |
| **apiRules.rules.path**            | No | APIRule | `/.*` | Specifies the path to the exposed Service.         |
| **apiRules.rules.path.accessStrategies.handler**            | Yes | APIRule | `allow` | Specifies one of the authenticators used: `oauth2_introspection`, `jwt`, `noop`, or `allow`.       |
| **apiRules.rules.path.accessStrategies.config.** | No | APIRule |  | Defines the handler used. It can be specified globally or per access rule.         |
| **apiRules.rules.path.accessStrategies.config.required_scope** | No | APIRule | | Defines the [limits](https://oauth.net/2/scope/) that the client specifies for an access request. In turn, the authorization server issues the access token in the defined scope. |
| **apiRules.rules.path.accessStrategies.config.jwks_urls** | No | APIRule | |  The URLs where ORY Oathkeeper can retrieve [JSON Web Keys](https://www.ory.sh/oathkeeper/docs/pipeline/authn/#jwt) from to validate the JSON Web Token.  |
| **apiRules.rules.path.accessStrategies.config.trustedIssuers** | No | APIRule | |  Sets a list of trusted token issuers. |
| **env.name**            | No | Function |  |  Specifies the name of the environment variable to export for the Function.  |
| **env.value**            | No | Function | |  Specifies the value of the environment variable to export for the Function.          |
| **env.valueFrom**            | No | Function | |  Specifies that you want the Function to use values either from a Secret or a ConfigMap. These objects must be stored in the same Namespace as the Function.     |
| **env.valueFrom.configMapKeyRef**            | No | Function | | Refers to the values from a ConfigMap that you want to use in the Function.         |
| **env.valueFrom.configMapKeyRef.Name**            | No | Function | | Specifies the name of the referred ConfigMap.        |
| **env.valueFrom.configMapKeyRef.Key**            | No | Function | | Specifies the key containing the referred value from the ConfigMap.       |
| **env.valueFrom.secretKeyRef**            | No | Function | | Refers to the values from a Secret that you want to use in the Function.        |
| **env.valueFrom.secretKeyRef.Name**            | No | Function | | Specifies the name of the referred Secret.         |
| **env.valueFrom.secretKeyRef.Key**            | No | Function | | Specifies the key containing the referred value from the Secret.               |

## Related resources

See the detailed descriptions of all related custom resources referred to in the `config.yaml`:

- [Function](./05-technical-reference/06-custom-resources/svls-01-function.md)
- [Subscription](./05-technical-reference/06-custom-resources/evnt-01-subscription.md)
- [APIRule](./05-technical-reference/06-custom-resources/apig-01-apirule.md)
