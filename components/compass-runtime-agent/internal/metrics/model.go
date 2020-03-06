package metrics

import "time"

type ClusterInfo struct {
	Resources []NodeResources `json:"resources"`
	Usage     []NodeMetrics   `json:"usage"`
}

type NodeResources struct {
	Name         string        `json:"nodeName"`
	InstanceType string        `json:"instanceType"`
	Capacity     ResourceGroup `json:"capacity"`
}

type NodeMetrics struct {
	Name                     string        `json:"nodeName"`
	StartCollectingTimestamp time.Time     `json:"startCollectingTimestamp"`
	Usage                    ResourceGroup `json:"usage"`
}

type ResourceGroup struct {
	CPU              string `json:"cpu"`
	EphemeralStorage string `json:"ephemeralStorage"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
}
