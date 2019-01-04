package shared

type ServiceInstance struct {
	Name                 string
	Environment          string
	ClassReference       ServiceInstanceResourceRef
	PlanReference        ServiceInstanceResourceRef
	PlanSpec             map[string]interface{}
	ClusterServicePlan   ClusterServicePlan
	ClusterServiceClass  ClusterServiceClass
	ServicePlan          ServicePlan
	ServiceClass         ServiceClass
	CreationTimestamp    int
	Labels               []string
	Status               ServiceInstanceStatus
	ServiceBindings      ServiceBindings
	ServiceBindingUsages []ServiceBindingUsage
	Bindable             bool
}

type ServiceInstanceResourceRef struct {
	Name        string
	DisplayName string
	ClusterWide bool
}

type ServiceBindings struct {
	Items []ServiceBinding
	Stats ServiceBindingStats
}

type ServiceBindingStats struct {
	Ready   int
	Failed  int
	Pending int
	Unknown int
}

type ServiceInstanceStatus struct {
	Type    string
	Reason  string
	Message string
}

const (
	ServiceInstanceStatusTypeRunning = "RUNNING"
)
