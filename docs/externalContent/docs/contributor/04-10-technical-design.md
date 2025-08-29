# Technical Design of Kyma API Gateway Operator

API Gateway Operator consists of two controllers that reconcile different CRs. To understand the reasons for using a single operator with multiple controllers instead of multiple operators, refer to the [Architecture Decision Record](https://github.com/kyma-project/api-gateway/issues/495).
API Gateway Operator has a dependency on [Istio](https://istio.io/) and [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper), and it installs Ory Oathkeeper itself.

Oathkeeper Deployment configures **PodAntiAffinity** to ensure that its Pods are evenly spread across all nodes and, if possible, across different zones. This guarantees High availability (HA) of the [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) installation.

The following diagram illustrates the APIRule reconciliation process and the resources created in the process:

![Kyma API Gateway Overview](../assets/operator-contributor-skr-overview.svg)

## APIGateway Controller

APIGateway Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md).

### Reconciliation
APIGateway Controller reconciles the APIGateway CR with each change. If you don't make any changes, the reconciliation process occurs at the interval of 10 hours.
APIGateway Controller reconciles only the oldest APIGateway CR in the cluster. It sets the status of other CRs to `Warning`.
If a failure occurs during the reconciliation process, the default behavior of the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) is to use exponential backoff requeue.

Before deleting the APIGateway CR, APIGateway Controller first checks if there are any APIRule or [Istio Virtual Service](https://istio.io/latest/docs/reference/config/networking/virtual-service) resources that reference the default Kyma [Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/) `kyma-system/kyma-gateway`. If any such resources are found, they are listed in the logs of the controller, and the APIGateway CR's status is set to `Warning` to indicate that there are resources blocking the deletion. If there are existing Ory Oathkeeper Access Rules in the cluster, APIGateway Controller also sets the status to `Warning` and does not delete the APIGateway CR.
The `gateways.operator.kyma-project.io/api-gateway-reconciliation` finalizer protects the deletion of the APIGateway CR. Once no more APIRule and VirtualService resources are blocking the deletion of the APIGateway CR, the APIGateway CR can be deleted. Deleting the APIGateway CR also deletes the default Kyma Gateway.

## APIRule Controller

APIRule Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the [APIRule CR](../user/custom-resources/apirule/04-10-apirule-custom-resource.md).
Additionally, the controller watches the [`api-gateway-config`](../user/custom-resources/apirule/04-15-api-rule-access-strategies.md) to configure the JWT handler.

APIRule Controller has a conditional dependency to APIGateway Controller in terms of the default APIRule domain. If you don't configure any domain in APIGateway CR, APIRule Controller uses the default Kyma Gateway domain as the default value for creating VirtualServices.

>**NOTE:** For now, you can only use the default domain in APIGateway CR. The option to configure your own domain will be added at a later time. See the [epic task](https://github.com/kyma-project/api-gateway/issues/130).

### Versions of APIRule CRDs
The latest version of the APIRules CRD is `v2`. The APIRule resource is stored and served as version `v1beta1`. APIRule `v2` is an exact copy of version `v2alpha1`, with the only difference being the version of the CRD.

### Reconciliation
APIRule Controller reconciles APIRule CR with each change. If you don't make any changes, the process occurs at the default interval of 30 minutes.
You can use the [API Gateway Operator parameters](../user/technical-reference/05-00-api-gateway-operator-parameters.md) to adjust this interval.
In the event of a failure during the reconciliation, APIRule Controller performs the reconciliation again after one minute.

The following diagram illustrates the reconciliation process of APIRule and the created resources:

![APIRule CR Reconciliation](../assets/api-rule-reconciliation-sequence.svg)

#### Reconciliation Processors
The APIRule reconciliation supports different processors that are responsible for validation and status handling as well as creating, updating, and deleting the resources in the cluster. 
The processor used is evaluated for each reconciliation of an APIRule and is determined by the configuration of the JWT handler in the `api-gateway-config` ConfigMap or the existence of the
annotation `gateway.kyma-project.io/original-version: v2alpha1` on the APIRule.

The processor is selected based on the following rules:
- If the handler in the `api-gateway-config` ConfigMap is set to `istio`, the APIRule reconciliation uses the `NewIstioReconciliation` in the [istio](../../internal/processing/processors/istio) package. 
- If the handler in the `api-gateway-config` ConfigMap is set to `ory`, the APIRule reconciliation uses the `NewOryReconciliation` in the [ory](../../internal/processing/processors/ory) package.
- If the annotation `gateway.kyma-project.io/original-version: v2alpha1` or `v2`  are present on the APIRule, the APIRule reconciliation uses the `NewReconciliation` in the [v2alpha1](../../internal/processing/processors/v2alpha1) package.

## Certificate Controller

Certificate Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the Secret `api-gateway-webhook-certificate` in the `kyma-system` namespace. This Secret contains the Certificate data required for the APIRule conversion webhook.

### Reconciliation
Certificate Controller reconciles a Secret CR with each change. If you don't make any changes, the process occurs at the default interval of 1 hour. This code verifies whether the Certificate is currently valid and will not expire within the next 14 days. If the Certificate does not meet these criteria, it is renewed. In the event of a failure during the reconciliation, Certificate Controller performs the reconciliation again with the predefined rate limiter.

## RateLimit Controller

RateLimit Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the [RateLimit CR](../user/custom-resources/ratelimit/04-00-ratelimit.md).

### Reconciliation
RateLimit Controller reconciles the RateLimit CR with each change. If you don't make any changes, the process occurs at the default interval of 30 minutes.
In the event of a failure during the reconciliation, RateLimit Controller performs the reconciliation again with the predefined rate limiter.
