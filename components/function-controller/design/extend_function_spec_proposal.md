# Extending CRD spec for Serverless v1alpha2

## Summary

The current Serverless API allows for limited configuration of the generated Function's deployment. 
Currently, users can only use ENVs to pass 3rd party service credentials but it doesn't allow for volume-mounted Secrets (which become the industry standard for [service bindings](https://servicebinding.io/application-developer/)).
Moreover, users have no control over the annotations applied on the function runtime pod. This excludes Function's Pods from features enabled by annotations (i.e., custom log parsers via `fluentbit.io/parser: my-regex-parser`).



## Motivation

Give Serverless users the ability to:
- Configure volume-mounted Secrets for Function's subresources.
- Configure labels and annotations for the Function's runtime Pod.

### Goals

- Add more flexibility to the Serverless API - enable volume-mounted Secrets and Pod annotations.
- Organise spec attributes belonging to runtime and build-time configuration (?)
- Propose sample Function CR visualising different variants

## Discussion points

### Runtime, build-time separation

Since Function CR is managing two workloads (Deployment for runtime and Job for build-time) we need to separate them in the spec in order to make it clear where the mounts and annotations belong. 

We can either make a clear separation, i.e.:

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

Alternatively, we could promote runtime fields to the root (as they belong to kind: Function) and extract only the build-time fields 

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

Under the hood, the Secret mount becomes a volume mount in the runtime Pod.
We could:

A) expose the k8s volume mount spec in the Function spec

Pros: 
 - Generic solution - Allows mounting anything: Secrets, ConfigMaps, any volumes
 - very easy to achieve (rewriting from the Function spec to Pod template spec)

Cons: 
 - Does not represent the service binding use case. User needs to translate service bindings into volume mounts by themselves
 - Less compact (elegant) to configure requested service binding use case
 - Noone has requested it yet

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
 - Less confusing (as volume mounts confuse Serverless Functions are considered stateless and should not claim any persistence volumes )
 - Enables using service binding natively with dedicated SDKs (i.e @sap/xsenv) - [related read](https://blogs.sap.com/2022/07/12/the-new-way-to-consume-service-bindings-on-kyma-runtime/)

Cons: 
 - Not allows mounting anything besides Secrets

A) and B) are not exclusive

We could separate those cases. (See last 'compromise' option)


### Metadata for Function's Pod

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
- dif1: move some fields from `.spec` to the new struct `.spec.runtimeSpec` to clearly distinguish fields desired to be used in the `build` and `running` phases. For example in `Option 1` users may have questions after seeing `.spec.env` and `.spec.build.env` fields for example "is .spec.env dedicated for the running Function's Pod? Would the field be merged with .spec.build.env for the building job?"
- dif2: rename the `.spec.profile` to the `.spec.resourcesProfile` to make this field more intuitive
- dif3: this solution is simple but not as simple as `Option 1`

>NOTE: the main idea is to close a specific configuration in a field that represents the specific phase of the Function's lifecycle. It would be intuitive and easy to understand for a user that the `.spec.build` field contains configuration for the building phase and `.spec.runtimeSpec` for the running phase.

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

This option exposes runtime pod configuration over the build Pod because the runtime Pod is the final result, and the build is a transient phase.
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


### Final version - the compromise

 - clearly separate configuration of the build stage in the spec (treating `build` stage as second class citizen and keeping the main spec dedicated to the more important runtime stage).
 - add convenient way to mount Secrets (w/o polluting function API with dependencies to service bindings)
 - volume mounts as a separate, more advanced case (implemented once requested) 
 - don't group labels and annotations under metadata. 

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function
spec: #Contains spec of Function and run stage (deployments, hpa)
  runtime: nodejs16
  source:
    ...
  replicas: 1
  scalingConfig:
    min: 1
    max: 2

  labels:
    app: my-app
  annotations:
    fluentbit.io/parser: my-regex-parser
    istio-injection: enabled

  secretMounts: 
  - secretName: my-secret
    mountPath: "/foo" #required.. no assumptions/validations towards SERVICE_BINDING_ROOT env value

  - secretName: my-redis-secret
    mountPath: "/bar" # this matches SERVICE_BINDING_ROOT env value. Its a soft indication that redis will be consumed as service binding 

  profile: S / M / L / XL / ... #optional
  resources: # optional... required if spec.profile is empty
    limits: ... #if profile empty
    requests: ... #if profile empty
​
  env:
  - name: SERVICE_BINDING_ROOT #set explicitely by user if he wants to use a specialised library that expects the ENV (and allows consumption of mounted secrets as service bindings)
    value: /bar
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

  # optional 
  build: #Contains spec of build stage : build is a "second class citizen" here. Users should make no assumptions that anything from main spec is inherited here (i.e ENVs or secretMounts)
    labels: 
    annotations: 
    profile: S / M / L / XL / ... #optional
    resources: # optional... required if spec.profile is empty
      limits: ... #if profile empty
      requests: ... #if profile empty


```

### Precedence, defaulting and validation

`profile` takes precedence over `resources`. If the profile field is not set there should be no defaulting happening for the resources. The controller should fill the pod template resources according to the selected profile preset.
If the profile field is set to "Custom", the user must then (and only then) set the values for resources manually.
Custom resource values together with non-custom profile should be rejected by the validation webhook.


### Labels and annotations

Function CR may have own labels and annotations as defined in it's metadata section. Those labels and annotations are automatically inherited by the resources managed directly by the function CR.
This direct, first-line inheritance of labels and annotations apply to:
 - Deployment
 - Job
 - HPA
 - ConfigMap

The same labels and annotation are NOT inherited by the Pods controlled by Deployment and Job (second-line inheritance of labels and annotation doesn't apply).
In order to control the labels and annotations on the runtime and build-time Pods user must define those in the spec:

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
  labels:
    app.kubernetes.io/name: my-function # <-- this label will be inherited by deployment, job, HPA and config map
spec: 
  ...
  labels:
    app: my-app # <-- this label will be used in Deployment's PodTemplate and will be applied on the function's runtime pod 
  annotations:
    fluentbit.io/parser: my-regex-parser # <-- those annotations will be used in Deployment's PodTemplate and will be applied on the function's runtime pod
    istio-injection: enabled

  ...  
  build:
    labels: # <-- those labels will be used in Job's PodTemplate and will be applied on the function's build-time pod
    annotations: # <-- those annotations will be used in Job's PodTemplate and will be applied on the function's build-time pod
