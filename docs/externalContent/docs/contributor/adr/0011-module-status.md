# Module Status Handling

## Status
Accepted

## Context

Currently, the API Gateway module sets the state `Processing` on the APIGateway custom resource during reconciliation or installation.
From the monitoring and Kubernetes API standpoint, this is not easy to handle properly. This is because it is not possible to observe the module's readiness, signified by the module CR being in the `Ready` state, without considering the periodic switch to the `Processing` state. This ADR proposes changes to handling the state of the module CR.

## Decision                                                                                                                                                                                                                                                                                                                                                                                                                             
The team decided to improve the state transition logic, with the `Processing` state only being set when the module is installed or reconfigured and there is a downtime possibility.
This decision was discussed with lead Kyma Architect. As general guidance, the processing state should only be set in case the module is possibly NOT ready. As is the case for the API Gateway module, this state should not occur unless the module's user changes the configuration (for example, disables default Kyma Gateway). The module should almost always be considered `Ready`.

## Consequences
As the initial installation could be considered the most important moment to observe the `Processing` state,
this solution is a good compromise between completely removing the state and keeping it in the API. However, the logic for handling
the `Processing` state cannot be entirely removed from the module.

## Alternative Solutions
See the alternative solutions that the team proposed and discussed but chose not to pursue.

### Solution 1

Proposal: Remove the `Processing` state entirely from the module's API.

Consequences: From a technical standpoint, this would be a breaking change, as there might be users relying on the `Processing` state,
for example, to determine if the module is being installed or not.
However, as the purpose of the `Processing` state is not clear, it might be a good idea to remove it entirely.

### Solution 2

Proposal: Perform a `soft` removal of the `Processing` state. 
This would mean that this state is still present in the API, but will never be set by the module.

Consequences: This approach would be less disruptive than the first solution, as the `Processing` state would remain in the API. 
However, the installation process would no longer be indicated by this state.

### Solution 3

Proposal: Improve the state transition logic, and set the `Processing` state when the module is installed or the user changes ANY configuration of the APIGateway custom resource.

Consequences: This solution would be similar to the one in the decision. Since not all configuration changes necessarily cause downtime, setting the state to `Processing` is not always mandatory.
