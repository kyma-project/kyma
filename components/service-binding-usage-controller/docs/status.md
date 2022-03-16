# Service Binding Usage status

The Service Binding Usage status ensures that the [Service Binding Usage](https://kyma-project-old.netlify.app/docs/components/service-catalog/#custom-resource-service-binding-usage) does not inject a Secret into the application without user's knowledge.

The status stores the latest condition which represents the current state of the Service Binding Usage.
The Service Binding Usage conditions are updated every time when the controller processes the Service Binding Usage. 
Find the list of all conditions and their descriptions in [this](../internal/controller/status/usage.go) file.

When the required Service Binding does not exist and the Service Biding Usage cannot inject a Secret into a given application, the Service Binding Usage Controller retries for a specified number of times and after that the Service Binding Usage status is marked as failed. The Controller attempts to reprocess the Service Binding Usage during synchronization with the informer. The time period of the informer synchronization is set in the [main.go](../cmd/controller/main.go) file under the **informerResyncPeriod** parameter. 

If the resource from the [Usage Kind](https://kyma-project-old.netlify.app/docs/components/service-catalog/#custom-resource-usage-kind) related to the successful Service Biding Usage is removed, its state is processed after time specified in the **informerResyncPeriod** parameter. If the Service Binding Usage fails due to the lack of Service Binding, its state is processed as soon as successful Service Binding appears.

