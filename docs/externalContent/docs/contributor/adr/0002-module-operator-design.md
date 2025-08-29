# Module Operator Design

## Status
Accepted

## Context
The current API Gateway operator consists of a controller that reconciles APIRule CRs. This controller also provides the ability to configure startup parameters. Some of these parameters affect the creation of resources, for example, the default domain changes how an Istio Virtual Service is created if no domain is set in the host field of the APIRule.

The original idea was to create a new operator that reconciles an APIGateway CR. This operator has a controller that deploys the existing API Gateway operator and sets the startup parameters that can be configured in the APIGateway CR. Additionally, the new operator installs Ory Oathkeeper and reconciles other resources such as Istio Gateways.

An alternative approach is to use one operator with two controllers (API Gateway controller and API Rule controller). In this case, a new controller is added to the current API Gateway operator, so each controller would reconcile separate CRs (APIGateway CR and APIRule CR). To solve the problem that the startup configuration of the API Rule controller is set in the APIGateway CR and reconciled by another controller, the startup parameters should be replaced by static default values.

Unfortunately, it is not possible to replace the start parameter for the default domain with a static default value. The domain can be read from the cluster configuration, but it should be possible to configure the default domain via the API gateway CR. 
Since controllers should be self-contained and a controller should not configure or manage another controller, the idea is to rely on the default reconciliation of APIRules to reconcile an updated default domain. This means that a default domain can be configured in the APIGateway CR and during the default reconciliation of the APIRule (which occurs after one hour), this default domain is read and taken into account during the reconciliation. 

## Decision
The decision was made to use one operator with two controllers. As we are aiming for a release of the API Gateway module as early as possible, the one operator approach allows us to proceed faster, because of the simpler development, deployment and release process, since we would need an additional repository with pipelines and a separate release process. Also after the initial release the processes will be easier with one operator.

![Diagram of API Gateway Operator with two controllers](https://github.com/kyma-project/api-gateway/assets/11753933/9873b3b7-1d8f-4ddd-89d8-aa56c8161b1e)

The default domain will be configured by the APIGateway CR and the APIRule controller will update the APIRules on change of the APIRule or during the next default reconciliation. We assume that changing the default domain is a task that is performed very rarely, which is why we accept the risk of a comparatively long tuning time.

## Consequences
We can use the existing api-gateway repository for the implementation of the new controller. We need to adapt this repository to be compliant with the modularisation concept and add the new controller. Since we already have everything set up and configured it's only adjusting existing configuration.

There are some tradeoffs that we have to accept. Although we want to separate the controllers, the controllers will be loosely coupled through the API gateway CR. The API Rule Controller must know the API Gateway CR to determine the default domain. However, the APIRule Controller should not validate this, but treat it as optional and reconcile the API Rule.
Furthermore, the first version of the operator will perform the installation of Ory Oathkeeper in the API Gateway Controller. Since certain APIRules (e.g. with a configured oauth handler) require Ory Oathkeeper resources, a configuration of the API Gateway CR is necessary in this case, otherwise the ORY resources for the API Rule cannot be created.

Having both controllers in the same repository means that we have to be very careful that the controllers remain separate and do not refer directly to each other.

Another consequence is that we need to implement the modular operator in a long-lived branch, since we still want to be able to release the APIGateway controller in its current state. Nevertheless, we want to merge this branch with the main branch as soon as possible, since we know about the disadvantages of long-lived branches.