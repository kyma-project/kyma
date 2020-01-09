# ServiceBindingUsage status

The ServiceBindingUsage status ensures that the [ServiceBindingUsage](https://kyma-project.io/docs/components/service-catalog#custom-resource-service-binding-usage) does not inject a Secret into the application without user's knowledge.

The status stores the latest condition which represents the current state of the ServiceBindingUsage.
The ServiceBindingUsage conditions are updated every time when the controller processes the ServiceBindingUsage. 
Find the list of all conditions and their descriptions in [this](../internal/controller/status/usage.go) file.

When the required ServiceBinding does not exist and the ServiceBidingUsage cannot inject a Secret into a given application, the ServiceBindingUsage controller retries for a specified number of times and after that the ServiceBindingUsage status is marked as failed. The controller attempts to reprocess the ServiceBidingUsage during synchronization with the informer. The time period of the informer synchronization is set in the [main.go](../cmd/controller/main.go) file under the **informerResyncPeriod** parameter. 

If the resource from the [UsageKind](https://kyma-project.io/docs/components/service-catalog/#custom-resource-usage-kind-sample-custom-resource) related to the successful ServiceBidingUsage is removed, its state is processed after time specified in the **informerResyncPeriod** parameter. If the ServiceBidingUsage fails due to the lack of ServiceBiding, its state is processed as soon as successful ServiceBiding appears.

