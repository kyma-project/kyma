# API Gateway Operator Parameters

You can configure APIGateway Controller and APIRule Controller using various parameters. This document contains all the options.

## Reconciliation Interval

### APIGateway
Kyma API Gateway Operator reconciles the APIGateway custom resource (CR) every 10 hours or whenever it is changed.

### APIRule
By default, API Gateway Operator reconciles APIRule CRs every 60 minutes or whenever an APIRule CR is changed. You can adjust this interval by modifying the operator's parameters. For example, you can set the **-reconciliation-interval** parameter to `120s`.

## Configuration Parameters

| Name                          | Required | Description                                                                                                            | Example values |
|-------------------------------|:--------:|------------------------------------------------------------------------------------------------------------------------|----------------|
| **metrics-bind-address**      |    NO    | The address the metric endpoint binds to.                                                                              | `:8080`        |
| **health-probe-bind-address** |    NO    | The address the probe endpoint binds to.                                                                               | `:8081`        |
| **leader-elect**              |    NO    | Enable leader election for API Gateway Operator. Enabling this ensures there is only one active APIGateway Controller. | `true`         |
| **rate-limiter-burst**        |    NO    | Indicates the burst value for the controller's bucket rate limiter.                                                    | 200            |
| **rate-limiter-frequency**    |    NO    | Indicates the controller's bucket rate limiter frequency, signifying no. of events per second.                         | 30             |
| **failure-base-delay**        |    NO    | Indicates the failure-based delay for rate limiter.                                                                    | `1s`           |
| **failure-max-delay**         |    NO    | Indicates the maximum failure delay for rate limiter.                                                                  | `1000s`        |
| **reconciliation-interval**   |    NO    | Indicates the time-based reconciliation interval of APIRule.                                                           | `1h`           |
| **migration-interval**        |    NO    | Indicates the time-based migration interval of APIRule.                                                                | `1m`           |