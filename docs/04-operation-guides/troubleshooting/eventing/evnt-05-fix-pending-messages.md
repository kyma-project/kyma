---
title: Published events are pending in the stream
---

## Symptom

You publish events, but some of them are not received by the subscriber and stay pending in the stream.

## Cause

When NATS EventingBackend has more than 1 replica and the `Clustering` property on the NATS Server is enabled,
a leader-election is taking place on the stream and consumer levels (see [NATS Documentation](https://docs.nats.io/running-a-nats-service/configuration/clustering/jetstream_clustering)).
Once the leader is elected, all the messages are being replicated across the replicas.

Sometimes replicas can go out-of-sync with the other replicas.
As a result of this, messages on some consumers can stop being acknowledged and start piling up in the stream.  

## Remedy

There are two ways of how to fix the "broken" consumers with pending messages. You will need either to trigger a leader reelection either on the consumers
with pending messages or on the stream level.

### Trigger Consumer leader election
First, you need to find out which consumer(s) have pending messages. For that you need the latest version of NATS cli installed on your machine.
You can find the broken consumer in two ways: by using Grafana dashboard or by directly using the NATS cli command.

#### Find the broken consumers using Grafana dashboard

1. [Access and Expose Grafana](../../security/sec-06-access-expose-grafana.md)
2. Find the NATS JetStream Dashboard and check the pending messages
   ![Pending consumer](./../../assets/grafana_pending_consumer.png)
3. Find the consumer with pending messages and encode it as md5 hash:
```bash
echo -n "tunas-testing/test-noapp3/kyma.sap.kyma.custom.noapp.order.created.v1" | md5
```
this shell command results in `6642d54a92f357ba28280b9cb609e79d`.
4. Then, you need to find consumer's leader:
```bash
nats consumer info sap 6642d54a92f357ba28280b9cb609e79d
```

```bash
Information for Consumer sap > 6642d54a92f357ba28280b9cb609e79d created 2022-10-24T15:49:43+02:00

Configuration:

                Name: 6642d54a92f357ba28280b9cb609e79d
         Description: tunas-testing/test-noapp3/kyma.sap.kyma.custom.noapp.order.created.v1
          ...

Cluster Information:

                Name: eventing-nats
              Leader: eventing-nats-1 # that's what we need
             Replica: eventing-nats-0, current, seen 0.96s ago
             Replica: eventing-nats-2, current, seen 0.96s ago
```
You can see, that its leader is the `eventing-nats-1` replica.

#### Find the broken consumers using the NATS cli
If you have NATS cli installed on your machine, you can simply run this shell script:
   ```bash
   for consumer in $(nats consumer list -n $stream_name)
   do
     nats consumer info sap $consumer -j |jq -c '{name: .name, pending: .num_pending, leader: .cluster.leader} '
   done
   ```
this will output the following:
```bash
{"name":"6642d54a92f357ba28280b9cb609e79d","pending":25,"leader":"eventing-nats-1"}
{"name":"c74c20756af53b592f87edebff67bdf8","pending":0,"leader":"eventing-nats-0"}
```
here you can see, that the consumer `6bfe94b513b39fb348e97b959c632e28` has pending messages. The other one has no pending message and 
is successfully processing events. 

#### Trigger the consumer leader reelection
Now, when we know the name of the broken consumer and its leader, we can trigger the reelection:

1. You need to port-forward the leader replica and trigger the leader reelection for that broken consumer
```bash
kubectl port-forward -n kyma-system eventing-nats-1 4222  
```
2. Trigger the leader reelection:
```bash
nats consumer cluster step-down $stream_name 6642d54a92f357ba28280b9cb609e79d
```
After execution, you see the following message:
```yaml
New leader elected "eventing-nats-2"

Information for Consumer sap > 6642d54a92f357ba28280b9cb609e79d created 2022-10-24T15:49:43+02:00
```
You can check the consumer now and confirm that the pending messages started to be dispatched.

### Restart the NATS pods and trigger the stream leader reelection
Sometimes triggering the leader reelection on the broken consumers doesn't work. In that case you should try to trigger leader reelection on the stream level:

```bash
nats stream cluster step-down $stream_name
```
As result, you will see:
```bash
11:08:22 Requesting leader step down of "eventing-nats-1" in a 3 peer RAFT group
11:08:23 New leader elected "eventing-nats-0"

Information for Stream $stream_name created 2022-10-24 15:47:19

             Subjects: kyma.>
             Replicas: 3
              Storage: File
```




