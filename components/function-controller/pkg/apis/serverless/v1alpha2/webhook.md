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

If we want to deal with Kubernetes Labels we should create a separate issue for implementing them and then decide if we want to allow overriding them.
https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/







