---
title: Serverless limitations
---

## Controller limitations
Serverless controller does not serve time critical requests from user.
It reconciles Function Custom Resources (CR) stored at the K8S API Server and has no persistent state on it's own.

Serverless controller doesn't build or serve funtions using its own allocated runtime resources. It delegates this work to dedicated kubernetes workloads. It schedules (build-time)  jobs to build the function docker image and (runtime) pods to serve them once they are built. 
Refer to the [architecture](../../../05-technical-reference/00-architecture/svls-01-architecture.md) diagram for more details.

Having this in mind serverless controller does not have a requirement for horizontal scaling.
It scales vertically up to the `160Mi` of memory and `500m` of cpu time.

## Limitation for amount of functions
There is no upper limit of functions that can be run on kyma (similar to kubernetes workloads in general). Once user defines a function, its build jobs and runtime pods will be always requested by serverless controller but it's up to kubernetes to schedule them based on the available memory and cpu time on the k8s worker nodes. This is determined mainly by amount of k8s worker nodes (and the node auto-scaling capabilities) and  their computational capacity.

## Build phase limitation:
As described before, the build-time phase requires cpu-time and memory to schedule jobs.
The time necessary to build a function depends on:
 - selected [build profile](../../../05-technical-reference/svls-09-available-presets.md#build-jobs-resources) that determines the requested resources ( and their limits ) for the build phase 
 - number of and size of dependencies that need to be downloaded and bundled into the function image.
 - cluster nodes specification (see reference specification at the end of the article)

#### Node.js functions

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 24 sec    | 15 sec                              |
| 2 dependencies  | 26 sec    | 16 sec                              |


#### Python functions

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 30 sec    | 16 sec                              |
| 2 dependencies  | 32 sec    | 20 sec                              |

The shortest build time (the limit) is appriximate 15 seconds and requires no limitation of the build job resources and minimum number of dependencies that are pulled in during the build phase.

Running multiple function build jobs at once ( especially such with no limits) may drain the cluster resources. To mitigate such risk there is additional limit of 5 symultanous function builds. If a sixth one will be scheduled, it will be built once there will be a vacancy in the build queue.

This limitation is configurable via [`containers.manager.envs.functionBuildMaxSimultaneousJobs`](../../../05-technical-reference/00-configuration-parameters/svls-01-serverless-chart.md#configurable-parameters) 


## Runtime phase limitations
In the runtime the functions are serving user provided logic wrapped in WEB framework (`express` for Node.js and `bottle` for Python). Taking the user logic aside, those frameworks have limitation by themselves and depend on the selected [runtime profile](../../../05-technical-reference/svls-09-available-presets.md#functions-resources) and kubernetes nodes specification (see the reference specification at the end of this article).

The following describes response times of a selected runtime profiles for a "hello world" function requested at 50 requests/second. This describes the overhead of the serving framework itself. Any user logic added on top of that will ofcourse add extra miliseconds and needs to be profiled separately.

#### Node.js functions

|                               | XL     | L      | M      | S      | XS      |
|-------------------------------|--------|--------|--------|--------|---------|
| response time [avarage]       | ~13ms  | 13ms   | ~15ms  | ~60ms  | ~400ms  |
| response time [95 percentile] | ~20ms  | ~30ms  | ~70ms  | ~200ms | ~800ms  |
| response time [99 percentile] | ~200ms | ~200ms | ~220ms | ~500ms | ~1.25ms |

#### Python functions

|                               | XL     | L      | M      | S      |
|-------------------------------|--------|--------|--------|--------|
| response time [avarage]       | ~11ms  | 12ms   | ~12ms  | ~14ms  |
| response time [95 percentile] | ~25ms  | ~25ms  | ~25ms  | ~25ms  |
| response time [99 percentile] | ~175ms | ~180ms | ~210ms | ~280ms |


Obviously, the bigger runtime profile, the more resources are available to serve response quicker. Please consider those limits of the serving layer as a baseline - as this do not take your function logic into account.


### Scaling

Function runtime pods can be scaled horizontally from zero up to the limits of the available resources at the kubernetes worker nodes.
Please find the guide [here](../../../03-tutorials/00-serverless/svls-15-use-external-scalers.md).

## In-cluster docker registry limitations

Serverless comes with in-cluster docker registry for function images.
Because of it's [limitations](../../main-areas/serverless/svls-03-container-registries.md) this registry is only suitable for development:
 - registry capacity is limited to 20GB
 - there is no image lifecycle managment. Once an image gets stored in the registry it stays there until they are manually removed.

 >> Reference specification of the cluster used for profiling: 
 All measurements were done on a kubernetes with 5 AWS nodes of type `m5.xlarge worker` (4 CPU 3.1 Ghz x86_64 cores, 16 GiB memory) 