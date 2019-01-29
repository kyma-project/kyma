---
title: Bucket
type: Custom Resource
---

The `buckets.assetstore.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define the name of the cloud storage bucket for storing assets. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd buckets.assetstore.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that defines the storage bucket configuration.

```
apiVersion: assetstore.kyma-project.io/v1alpha1
kind: Bucket
metadata:
  name: my-bucket
  namespace: default
spec:
  location: "us-east-1"
  policy: `
     {
      "Version":"2012-10-17",
      "Statement":[
         {
            "Action":[
               "s3:GetObject"
            ],
            "Effect":"Allow",
            "Principal":{
               "AWS":[
                  "*"
               ]
            },
            "Resource":[
               "arn:aws:s3:::my-bucket/*"
            ],
            "Sid":""
         }
      ]
   }

```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the CR is available. |
| **spec.location** |    **NO**   | xxx  Specifies the [location](https://docs.minio.io/docs/golang-client-api-reference.html#MakeBucket) under which the bucket is created. If the field is empty, the bucket is created under the default location. |
| **spec.policy** |    **NO**   | There is no policy specified for the bucket by default. To change it, specify the policy in a string that is similar to `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::my-bucketname/*"],"Sid": ""}]}`. |


## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| Asset |  Provides the name of the storage bucket which the Asset CR refers to. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Asset Store |  Uses the Bucket CR for the storage bucket definition. |
