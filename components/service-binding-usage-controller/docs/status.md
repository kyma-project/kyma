# ServiceBindingUsage status

The ServiceBindingUsage status ensures that the ServiceBindingUsage does not inject a Secret into the application without the user's knowledge.

The status stores the latest condition which represents the current state of the ServiceBindingUsage.
The ServiceBindingUsage conditions are updated every time when the controller proceeds the ServiceBindingUsage.

For example, when the required ServiceBinding does not exist and the ServiceBidingUsage cannot inject a Secret into a given application, the ServiceBindingUsage Controller retries for a specified number of times.

If during a specified number of ServiceBidingUsage attempts the controller cannot find the appropriate resource related to the indicated application the status will be marked as failed.
The controller will attempt to reprocess the ServiceBidingUsage during synchronization with the informer. Period time in which the informer is synchronize is set in the main [file](../cmd/controller/main.go) under `informerResyncPeriod` parameter.
If resource from UsageKind related to the successful ServiceBidingUsage will be removed, its state will be processed after the time specified in the parameter `informerResyncPeriod`.
If the ServiceBidingUsage fails due to the lack of ServiceBiding, its state will be processed as soon as success ServiceBiding appears.

Find the list of all **conditions** and their descriptions in [this](../internal/controller/status/usage.go) file.
