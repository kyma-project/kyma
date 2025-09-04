# Kyma Eventing Metrics

Kyma Eventing provides various metrics, so you can monitor statistics and other information in real time.
The metrics follow the [Prometheus naming convention](https://prometheus.io/docs/practices/naming/).

## Metrics Emitted by Eventing Publisher Proxy

| Metric                                         | Description                                                                      |
| ---------------------------------------------- | :------------------------------------------------------------------------------- |
| **eventing_epp_backend_duration_milliseconds** | The duration of sending events to the messaging server in milliseconds           |
| **eventing_epp_event_type_published_total**    | The total number of events published for a given eventTypeLabel                  |
| **eventing_epp_health**                        | The current health of the system. `1` indicates a healthy system                 |
| **eventing_epp_requests_duration_seconds**     | The duration of processing an incoming request (includes sending to the backend) |
| **eventing_epp_requests_total**                | The total number of requests                                                     |

## Metrics Emitted by Eventing Manager

| Metric                                                    | Description                                                                                                                 |
| --------------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------- |
| **eventing_ec_event_type_subscribed_total**               | The total number of eventTypes subscribed using the Subscription CRD                                                        |
| **eventing_ec_health**                                    | The current health of the system. `1` indicates a healthy system                                                            |
| **eventing_ec_nats_delivery_per_subscription_total**      | The total number of dispatched events per subscription                                                                      |
| **eventing_ec_nats_subscriber_dispatch_duration_seconds** | The duration of sending an incoming NATS message to the subscriber (not including processing the message in the dispatcher) |
| **eventing_ec_subscription_status**                       | The status of a subscription. `1` indicates the subscription is marked as ready                                             |

### Metrics Emitted by NATS Exporter

The [Prometheus NATS Exporter](https://github.com/nats-io/prometheus-nats-exporter) also emits metrics that you can monitor. Learn more about [NATS Monitoring](https://docs.nats.io/running-a-nats-service/configuration/monitoring#jetstream-information).
