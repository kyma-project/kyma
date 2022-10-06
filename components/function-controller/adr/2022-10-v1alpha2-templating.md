# Function templating 
To  give user more control of workloads created by serverless controller, template concept was created.
The following decision has been made to templating concept.
 
## Metadata templating
This is the concept of dealing with metadata.
``` yaml
    kind: Function
    metadata: #Isolated, standalone field
        ...
    spec:
        template: #Old field that is being ignored
            labels
            annotations
        templates:
            functionPod:
                metadata: ...
                ...
            buildPod:
                metadata: ...
                ...
```
We concluded that Validation Webhook needs to check whether given labels or annotations are conflicting with our own and check if something overwrites our labels.

For example:
``` yaml
"serverless.kyma-project.io/function-name": "function-hello-world"
"serverless.kyma-project.io/managed-by": "function-controller", 
"serverless.kyma-project.io/resource": "deployment", 
"serverless.kyma-project.io/uuid": "98f05b9d-ecd1-4a70-96d6-5848ec4ed3a7",
```

### old field for function metadata
In v1alpha2 version we have `spec.template` field on root level which configures function pod metadata.
This field will be copied/moved to `spec.templates.functionPod.metadata`.

If we want to deal with Kubernetes Labels we should create a separate issue for implementing them and then decide if we want to allow overriding them.
https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

## Resources
After converting from "v1alpha1" we have duplicate fields and we have to decide how to handle them.
We use the "profile" ("preset") fields: "spec/resourceConfiguration/[function|build]/profile" (*P1*) or "metadata/labels/serverless.kyma-project.io/[function|build]-resources-preset" (*P2*) (deprecated).
For a custom (user defined) configuration we use the "resources" fields: "spec/templates/[functionPod|buildJob]/spec/resources" (*R1*) or "spec/resourceConfiguration/[function|build]/resources" (*R2*) (deprecated).
When the user fills both *P1* and *P2* (or *R1* and *R2*) we only take *P1* (*R1*). Otherwise we take the filled one.
The "profile" is propageted to the "resources". When we have both "profile" and "resources", the "profile" has a higher priority and overrides values in "resources".

Use cases:
* no profile, no resources -> we use default function resources and empty build resources,
* no profile, resources -> we use user definied resources,
* no profile, incomplete resources -> we use user definied resources (we don't fill the missing resources from the default profile),
* profile, no resources -> we use profile resources,
* profile, resources -> we use profile resources (we overwrite custom resources),
* profile, incomplete resources -> we use profile resources (we overwrite custom resources).

The `resourceConfiguration` object can be replaced by more specific object when we will create v1alpha3 to clean up deprecated fields.:
```yaml
resourceProfiles:
  build: L
  runtime: S
```

## Env
In v1alpha2 version we have `spec.env` field on root level which configures function pod envs.
This field will be copied/moved to `spec.templates.functionPod.spec.env`.

## VolumeMounts
This is new feature.
