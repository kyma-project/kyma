# ServiceBindingUsage status

The ServiceBindingUsage status ensures that the ServiceBindingUsage does not inject a Secret into the application without the user's knowledge.

The status stores the latest condition which represents the current state of the ServiceBindingUsage.
The ServiceBindingUsage conditions are updated every time when the controller proceeds the ServiceBindingUsage.

For example, when the required ServiceBinding does not exist and the ServiceBidingUsage cannot inject a Secret into a given application, the ServiceBindingUsage Controller retries for a specified number of times.

Find the list of all **conditions** and their descriptions in [this](../internal/controller/status/usage.go) file.
