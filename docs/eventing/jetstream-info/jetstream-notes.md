# NATs - JetStream

## Check deployment (nats operator vs helm charts)
* NATS Operator officially discourages to use nats-operator for new deployments. Also support for JetStreams by nats-operator is questionable.
> The recommended way of running NATS on Kubernetes is by using the Helm charts. If looking for JetStream support, this is supported in the Helm charts. The NATS Operator is not recommended to be used for new deployments.
([Source](https://github.com/nats-io/nats-operator#nats-operator))

### Deploy NATS with JetStream using Helm chart

>Reference: [JetStream with NATS Helm Chart](https://github.com/nats-io/k8s/tree/main/helm/charts/nats#jetstream)

1. Download latest nats helm chart from [here](https://github.com/nats-io/k8s/releases/).
2. Enable JetStream in `Values.yaml` file.
```
  jetstream:
    enabled: true
```
3. Configure clustering in `Values.yaml` file if needed. [More info](https://docs.nats.io/jetstream/clustering) on clustering with JetStream.
```
cluster:
  enabled: true
  # Cluster name is required, by default will be release name.
  # name: "nats"
  replicas: 3
  noAdvertise: false
```

## Current NATS workload works using Jetstream
--> Streams
* Create Stream and assign subjects to this stream. 
* Any event published to any of the subject will be recevied and stored by the stream.
* Streams define how messages are stored and retention duration/policies.
* Two storage types supported: Memory-based or File-based.
* Encryption at Rest supported for security, but can effect on performance.

--> Producers
* If you send a Request to the subject the JetStream server will reply with an acknowledgement that it was stored.

--> Consumers
* There are two types of consumers i.e. Push-based and Pull-based consumers.

## Extra Info
- Streams consume normal NATS subjects, any message found on those subjects will be delivered to the defined storage system.
- Streams support deduplication using a Nats-Msg-Id header and a sliding window within which to track duplicate messages.
- The [NATS Surveyor](https://github.com/nats-io/nats-surveyor) system has initial support for passing JetStream metrics to Prometheus, dashboards and more will be added towards final release.
- JetStream uses a NATS optimized RAFT algorithm for [clustering](https://docs.nats.io/jetstream/clustering).
- RAFT Groups: Meta group, Stream group, Consumer group.
- Each JetStream node must specify a server name and cluster name.
- The JetStream controllers (NACKs) allow you to manage NATS JetStream Streams and Consumers via K8S CRDs.

## Links
- [JetStream on K8s using Helm](https://docs.nats.io/nats-on-kubernetes/helm-charts#jetstream)
- [Model Deep Dive](https://docs.nats.io/jetstream/model_deep_dive)
- [NATS Golang client](https://github.com/nats-io/nats.go)



