# Extending CRD spec for Serverless v1alpha2

## Summary

The current Serverless API allows for limited configuration of the generated Function's deployment. 
Currently users can only use ENVs to pass 3rd party service credentials but it doesn't allow for volume mounted secrets (which becomes the industry standard for [service bindings](https://servicebinding.io/application-developer/)).
Moreover users have no control over the annotations that are applied on the function runtime pod. This excludes functions pods from features enabled by annotations (i.e custom log parsers via `fluentbit.io/parser: my-regex-parser`).



## Motivation

Give Serverless users the ability to:
- Configure volume mounted secrets for Function's subresources.
- Configure labels and annotations for the Function runtime pod.

### Goals

- Add more flexibility to the Serverless API - enable volume mounted secrets and pod annotations.
- Organise spec attributes belonging to runtime and build-time configuration (?)
- Propose sample Function CR visualising different variants
- Eventually, provide a rollout plan to implement the changes

### Stretch Goals
- Try to avoid a new API version specification

## Discussion points

### Runtime, build-time separation

Since Function CR is managing two workloads ( Deployment for runtime and Job for build-time ) we need to separate them in the spec in order to make it clear where do the mounts and annotations belong. 

We can either make a clear separation, i.e:

```yaml 
spec:
  source:
  runtimeSpec:
    # mounts
    # resources
    # envs
    # metadata
  buildSpec:
    # resources
    # envs (?)
    # metadata (?)
```

Alternatively we could promote runtime fields to the root (as the belong to kind: Function) and extract only the build-time fields 

```yaml 
spec:
  source:
  # mounts
  # resources
  # envs
  # metadata
  buildSpec:
    # resources
    # envs (?)
    # metadata (?)
```

### Mounts - own structure or k8s inherited

Under the hood the secret mount becomes a volume mount in the runtime pod.
We could :

A) expose the k8s volume mount spec in function spec

Pros: 
 - Generic solution - Allows mounting anything : secrets, config maps, any volumes
 - very easy to achieve (rewriting from function spec to pod teplate spec)

Cons: 
 - Does not represent the service binging use case. User needs to translate service bindngs into volume mounts by themselves
 - Less compact ( elegant ) to configure requested service binding use case
 - Noone requested it yet

```yaml 
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
spec:
  volumeMounts:
  - name: foo
    mountPath: "/etc/foo"
    readOnly: true
  volumes:
  - name: foo
    secret:
      secretName: mysecret
```
 B) focus on [service binding case](https://servicebinding.io/application-developer/).

```yaml 
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
spec:
  secretBindings: #serviceBindings will cause name clash with https://github.com/SAP/sap-btp-service-operator#service-binding
  - source: my-secret
    mountPath: /bar # optional mount path
  env:
  - name: SERVICE_BINDING_ROOT # default mount path for service bindings
    value: /foo
```
 Pros: 
 - Purpose focused  
 - Compact configuration - easy to adopt
 - Less confusing (as volume mounts cause confusion as serverless functions are considered stateless and should not claim any persistance volumes )
 - Enables using service binding natively with dedicated SDKs (i.e @sap/xsenv) - [related read](https://blogs.sap.com/2022/07/12/the-new-way-to-consume-service-bindings-on-kyma-runtime/)

Cons: 
 - Not allows to mount anything besides secrets

A) and B) are not exclusive

We could separate those cases. (See last 'compromise' option)


### Metadata for function pod

Define runtime labels and annotations on the root level or under a `metadata` field

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    ...
  annotations:
    ...
spec:
  metadata: 
    labels: ...
    annotations: ...
```
OR
```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    ...
  annotations:
    ...
spec:
  labels: ...
  annotations: ...
```

## Samples: 

### Option 1

- simplified mounts serving just for service binding purpose
- stretch: volume mounts could be added as a separate feature
- extract `build` object for any build-time specific config
- runtime labels and annotations on the spec root level

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec:
  runtime: nodejs16
  source:
    ...
  secretBindings: 
  - source: my-secret
    mountPath: "/foo" # optional.. read from SERVICE_BINDING_ROOT ENV 

  profile: S / M / L / XL / ... / Custom #optional
  resources: # optional... required if spec.profile==Custom
    limits: ... #if_custom
    requests: ... #if_custom
​  
  labels: 
    app: my-app
  annotations: 
    fluentbit.io/parser: my-regex-parser
    istio-injection: enabled
​
  env:
  - name: SERVICE_BINDING_ROOT
    value: /service_bindings
​​
  build:
    profile: S / M / L / XL / ... / Custom
    resources: # optional... required if spec.build.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom
    labels: #optional
    annotations: #optional
```

### Option 2

- almost the same as `Option 1` (same pros)
- dif1: move some fields from `.spec` to the new struct `.spec.runtimeSpec` to clearly distinguish fields desired to be used in the `build` and `running` phases. For example in `Option 1` users may have questions after seeing `.spec.env` and `.spec.build.env` fields for example "is .spec.env dedicated for the running function's pod? Would the field be merged with .spec.build.env for the building job?"
- dif2: rename the `.spec.profile` to the `.spec.resourcesProfile` to make this field more intuitive
- dif3: this solution is simple but not as simple as `Option 1`

>NOTE: the main idea is to close a specific configuration in a field that represents the specific phase of the function's lifecycle. It would be intuitive and easy to understand for a user that the `.spec.build` field contains configuration for the building phase and `.spec.runtimeSpec` for the running phase.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec:
  runtime: nodejs16
  source:
    ...
  runtimeSpec: # this name is not perfect
    serviceBindings: 
    - source: my-secret
      mountPath: "/foo" # optional.. read from SERVICE_BINDING_ROOT ENV 

    resourcesProfile: S / M / L / XL / ... / Custom #optional
    resources: # optional... required if spec.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom


    labels: 
      app: my-app
    annotations: 
      fluentbit.io/parser: my-regex-parser
      istio-injection: enabled
  
    env:
    - name: SERVICE_BINDING_ROOT
      value: /service_bindings

  buildSpec: # this name is not perfect
    resourcesProfile: S / M / L / XL / ... / Custom
    resources: # optional... required if spec.build.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom

```

### Option 3

This option expose runtime pod configuration over build pod because runtime pod is final result, the build is transient phase.
Additionaly:
- It allows to use full k8s volume api the similar way to k8s pods.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha3
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec:
  sources:
    inline:
      source: aaaa
      dependency: bbbb
  replicas: 1
  scalingConfig:
    min: 1
    max: 2

  resourcesProfile: S / M / L / XL / ... | (empty)-> resources field has to be filled 
  resources: #k8s limits and requests
  
  envs:
    - name: PASSWORD
      valueFrom:
        secretRef:
          name: mysvc-passwords
          key: password
    - name: EXTERNAL_API_URL
      valueFrom:
        configmapRef:
          name: mysvc-configuration
          key: URL
  volumeMounts:
    - name: config
      mountPath: /etc/config/
    - name: search-index
      mountPath: /etc/index
  volumes:
    - name: search-index
        nfs:
          path: /path-to-index
          readOnly: true
          server: localhost
    - name: config
      configmap:
        name: function-configuration
        items:
          - key: config
            path: config.yaml
  
  metadata:
    labels:
      app: my-app
    annotations:
      fluentbit.io/parser: my-regex-parser
      istio-injection: enabled
  
  #build can share the same configuration options as function: 
  # metadata, volumes, volumeMounts, envs, resourceProfile, resources, 
  build:
    metadata:
      labels:
        app: my-app
      annotations:
        fluentbit.io/parser: my-regex-parser
        istio-injection: enabled
    resourcesProfile: S / M / L / XL / ... | (empty)-> resources field has to be filled
    resources: #k8s limits and requests
    envs:
    - name: RUNTIME_CACHE_OFF
      value: true
    volumes:
      - name: private-deps-repo-configuration
        secret:
          secretName: private-repo
    volumeMounts:
      - name: private-deps-repo-configuration
        path: /etc/my-dep-resolver.config

```
### Option 4

Option presented before with separated configurations (templates) for build and function.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec:
  runtime: nodejs16
  source:
    ...
  resourceProfiles:
    function: S / M / L / XL / ... | (empty)-> resources field has to be filled
    build: S / M / L / XL / ... | (empty)-> resources field has to be filled
  replicas: 1
  scalingConfig:
    min: 1
    max: 2
  templates:
    functionPod:
      metadata: # labels and annotations only for function pod
        labels: 
          app: my-app
        annotations: 
          fluentbit.io/parser: my-regex-parser
          istio-injection: enabled
      spec: 
        resources: # optional... required if spec.resourceProfiles.function is empty
          limits: ... #if_custom
          requests: ... #if_custom
        env: # function pod envs
          - name: PASSWORD
            valueFrom:
              secretRef:
                name: mysvc-passwords
                key: password
          - name: EXTERNAL_API_URL
            valueFrom:
              configmapRef:
                name: mysvc-configuration
                key: URL
        volumeMounts: ...
          - name: config
            mountPath: /etc/config/
          - name: search-index
            mountPath: /etc/index
      volumes:
        - name: search-index
            nfs:
              path: /path-to-index
              readOnly: true
              server: localhost
        - name: config
          configmap:
            name: function-configuration
            items:
              - key: config
                path: config.yaml
    buildPod:
      metadata: ...# labels and annotations only for build pod
      spec: 
        resources: # optional... required if spec.resourceProfiles.build is empty
          limits: ... #if_custom
          requests: ... #if_custom
        env: ...
        volumeMounts: ...
      volumes: ...
```


### Compromise

 - clearly separate build and runtime stages in spec
 - add simple, [purpose focused](https://blogs.sap.com/2022/07/12/the-new-way-to-consume-service-bindings-on-kyma-runtime/) `secretBindings` 
 - volume mounts separated to a different place, for more advanced case (implemented once requested) 
 - use metadata under buildSpec and runtimeSpec

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec:
  runtime: nodejs16
  source:
    ...
  replicas: 1
  scalingConfig:
    min: 1
    max: 2

  runtimeSpec:

    metadata: #optional
      labels: 
        app: my-app
      annotations: 
        fluentbit.io/parser: my-regex-parser
        istio-injection: enabled

    secretBindings: 
    - source: my-secret
      mountPath: "/foo" # optional.. read from SERVICE_BINDING_ROOT ENV 

    profile: S / M / L / XL / ... / Custom #optional
    resources: # optional... required if spec.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom
​  
​
    env:
    - name: SERVICE_BINDING_ROOT
      value: /service_bindings
    - name: MODE
      value: modeA

    # volumeMounts:  (add when requested by users)
    # - name: config
    #   mountPath: /etc/config/
    # - name: search-index
    #   mountPath: /etc/index

    # volumes:
    # - name: search-index
    #     nfs:
    #       path: /path-to-index
    #       readOnly: true
    #       server: localhost
    # - name: config
    #   configmap:
    #     name: function-configuration
    #     items:
    #       - key: config
    #         path: config.yaml
​​
  buildSpec:
    metadata: #optional
      labels: 
      annotations: 
    profile: S / M / L / XL / ... / Custom
    resources: # optional... required if spec.build.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom


```

### Precedence, defaulting and validation

`Profile` takes precedence over `resources`. If profile field is not set to "Custom" there should be no defaulting happening for the resources. Controller should fill the pod template resources accoriding to the selected profile preset.
If profile field is set to "Custom", the user must then ( and only then ) set the values for resources manualy.
Custom resource values together with non-custom profile should be rejected by the validation webhook.
