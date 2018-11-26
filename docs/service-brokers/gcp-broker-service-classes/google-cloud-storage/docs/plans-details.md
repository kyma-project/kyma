---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Google Cloud Storage plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new Google Cloud Storage Bucket. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **billing** | `object` | The bucket's billing configuration. | NO | - |
| **billing.requesterPays** | `boolean` | If set to `true`, Requester Pays is enabled for this bucket. | NO | - |
| **bucketId** | `string` | Must be unique across Cloud Storage. Must contain only lowercase letters, numbers, dashes, underscores, and dots. Must start and end with an alphanumeric. Must be 3-63 characters long. | YES | - |
| **cors** | `array` | The bucket's Cross-Origin Resource Sharing (CORS) configuration. | NO | - |
| **cors.maxAgeSeconds** | `integer` | The value, in seconds, to return in the  Access-Control-Max-Age header used in preflight responses. | NO | - |
| **cors.method** | `array` | The list of HTTP methods on which to include CORS response headers, (GET, OPTIONS, POST, etc) Note: \"*\" is permitted in the list of methods, and means \"any method\". | NO | - |
| **cors.origin** | `array` | The list of Origins eligible to receive CORS response headers. Note: \"*\" is permitted in the list of origins, and means \"any Origin\". | NO | - |
| **cors.responseHeader** | `array` | The list of HTTP headers other than the simple response headers to give permission for the user-agent to share across domains. | NO | - |
| **defaultEventBasedHold** | `boolean` | Defines the default value for Event-Based hold on newly created objects in this bucket. Event-Based hold is a way to retain objects indefinitely until an event occurs, signified by the hold's release. After being released, such objects will be subject to bucket-level retention (if any). One sample use case of this flag is for banks to hold loan documents for at least 3 years after loan is paid in full. Here bucket-level retention is 3 years and the event is loan being paid in full. In this example these objects will be held intact for any number of years until the event has occurred (hold is released) and then 3 more years after that. Objects under Event-Based hold cannot be deleted, overwritten or archived until the hold is removed. | NO | - |
| **defaultObjectAcl** | `array` | Default access controls to apply to new objects when no ACL is provided. For more information, see the **ObjectAccessControl properties** section. | NO | - |
| **encryption** | `object` | Encryption configuration used by default for newly inserted objects, when no encryption config is specified. | NO | - |
| **encryptiondefaultKmsKeyName** | `string` | A Cloud KMS key that will be used to encrypt objects inserted into this bucket, if no encryption method is specified. Limited availability; usable only by enabled projects. | NO | - |
| **labels** | `object` | To organize your project, add arbitrary labels as key/value pairs to Cloud Spanner. Use labels to indicate different elements, such as environments, services, or teams. | NO | - |
| **lifecycle** | `object` | The bucket's lifecycle configuration. See lifecycle management for more information. | NO | - |
| **location** | `string` |  Determines where the Storage Bucket data is stored. These are the possible values of this parameter: `US`, `EU`, `ASIA`, `northamerica-northeast1`, `us-central1`, `us-east1`, `us-east4`, `us-west1`, `southamerica-east1`, `europe-west1`, `europe-west2`, `europe-west3`, `europe-west4`, `asia-east1`, `asia-northeast1`, `asia-south1`, `asia-southeast1`, `australia-southeast1`  | YES | `US` |
| **logging** | `object` | The bucket's logging configuration, which defines the destination bucket and optional name prefix for the current bucket's logs. | NO | - |
| **logging.logBucket** | `string` | The destination bucket where the current bucket's logs should be placed. | NO | - |
| **logging.logObjectPrefix** | `string` | A prefix for log object names. | NO | - |
| **storageClass** | `string` | The Cloud Storage bucket's default storage class. The possible values are: `MULTI_REGIONAL`, `REGIONAL`, `STANDARD`, `NEARLINE`, `COLDLINE`, `DURABLE_REDUCED_AVAILABILITY`| NO | `STANDARD` |
| **updated** | `string` | The modification time of the bucket in RFC 3339 format. | NO | - |
| **versioning** | `object` | The bucket's versioning configuration. | NO | - |
| **versioning.enabled** | `boolean` | While set to true, versioning is fully enabled for this bucket. | NO | - |
| **website** | `object` | The bucket's website configuration, controlling how the service behaves when accessing bucket contents as a web site. See the Static Website Examples for more information. | NO | - |
| **website.mainPageSuffix** | `string` | If the requested object path is missing, the service will ensure the path has a trailing '/', append this suffix, and attempt to retrieve the resulting object. This allows the creation of index.html objects to represent directory pages. | NO | - |
| **website.notFoundPage** | `string` | If the requested object path is missing, and any mainPageSuffix object is missing, if applicable, the service will return the named object from this bucket as the content for a 404 Not Found result. | NO | - |

### ObjectAccessControl properties
These are the properties of the **ObjectAccessControl** parameter:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **entity** | `string` |The entity holding the permission, in one of the following forms: \n- user-userId \n- user-email \n- group-groupId \n- group-email \n- domain-domain \n- project-team-projectId \n- allUsers \n- allAuthenticatedUsers Examples: \n- The user liz@example.com would be user-liz@example.com. \n- The group example@googlegroups.com would be group-example@googlegroups.com. \n- To refer to all members of the Google Apps for Business domain example.com, the entity would be domain-example.com. Required:  `storage.defaultObjectAccessControls.insert`, `storage.objectAccessControls.insert` | NO | - |
| **role** | `string` | The access permission for the entity. Required:  `storage.defaultObjectAccessControls.insert`, `storage.objectAccessControls.insert` | NO | - |

### Lifecycle properties

These are the properties of the **Lifecycle** parameter:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **rule** | `object` | A lifecycle management rule, which is made of an action to take and the condition(s) under which the action will be taken. | NO | - |
| **rule.action** | `object` | The action to take. | NO | - |
| **rule.action.storageClass** | `string` | Target storage class. Required iff the type of the action is SetStorageClass. | NO | - |
| **rule.action.type** | `string` | Type of the action. Currently, only Delete and SetStorageClass are supported. | NO | - |
| **rule.condition** | `object` | The condition(s) under which the action will be taken. | NO | - |
| **rule.condition.age** | `integer` | Age of an object (in days). This condition is satisfied when an object reaches the specified age. | NO | - |
| **rule.condition.createdBefore** | `string` | A date in RFC 3339 format with only the date part (for instance, \"2013-01-15\"). This condition is satisfied when an object is created before midnight of the specified date in UTC. | NO | - |
| **rule.condition.isLive** | `boolean` | Relevant only for versioned objects. If the value is true, this condition matches live objects; if the value is false, it matches archived objects. | NO | - |
| **rule.condition.matchesStorageClass** | `string` | Objects having any of the storage classes specified by this condition will be matched. Values include MULTI_REGIONAL, REGIONAL, NEARLINE, COLDLINE, STANDARD, and DURABLE_REDUCED_AVAILABILITY. | NO | - |
| **rule.condition.numNewerVersions** | `integer` | Relevant only for versioned objects. If the value is N, this condition is satisfied when there are at least N versions (including the live version) newer than this version of the object. | NO | - |


## Update parameters:

The update parameters are the same as the provisioning parameters.

## Binding parameters:

Binding grants the provided service account access to the Cloud Storage Bucket. Optionally, a new service account can be created and given access to the Cloud Storage Bucket. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Create a new service account for GCS binding. | NO | `false` |
| **roles** | `array` | The list of Cloud Storage roles for this binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/storage.objectCreator`, `roles/storage.objectViewer`, `roles/storage.objectAdmin`, `roles/storage.admin`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |
