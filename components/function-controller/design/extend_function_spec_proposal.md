# Extending CRD spec for Serverless v1alpha2

## Summary

The current Serverless API allows for limited configuration of the generated Function's deployment. 
Currently users can only use ENVs to pass 3rd party service credentials but it doesn't allow for volume mounted secrets (which becomes the industry standard for [service bindings](https://servicebinding.io/application-developer/)).
Moreover users have no control over the annotations that are applied on the function runtime pod. This excludes functions pods from features enabled by annotations (i.e custom log parsers via `fluentbit.io/parser: my-regex-parser`).



## Motivation

Give Serverless users the ability to:
- Configure volume mounted secrets (and config maps) for Function's subresources.
- Configure labels and annotations for the  Function runtime pod.

### Goals

- Add more flexibility to the Serverless API - enable volume mounted secrets and pod annotations.
- Organise spec attributes belonging to runtime and build-time configuration (?)
- Provide a rollout plan to implement this functionality in small increments and avoid the need to roll out a new API version

### Non-Goals
- Define a new API version specification

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
 - expose the k8s volume mount spec  in function spec
 - cover k8s volume mount spec via oor own [specialised](https://servicebinding.io/application-developer/) spec

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

```yaml 
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
spec:
  serviceBindings:
  - source: my-secret
  env:
  - name: SERVICE_BINDING_ROOT 
    value: /foo
```

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
  serviceBindings: 
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
​
​
  build:
    profile: S / M / L / XL / ... / Custom
    resources: # optional... required if spec.build.profile==Custom
      limits: ... #if_custom
      requests: ... #if_custom
​
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
### Option 4

### Precedence, defaulting and validation

`Profile` takes precedence over `resources`. If profile field is not set to "Custom" there should be no defaulting happening for the resources. Controller should fill the pod template resources accoriding to the selected profile preset.
If profile field is set to "Custom", the user must then ( and only then ) set the values for resources manualy.
Custom resource values together with non-custom profile should be rejected by the validation webhook.
