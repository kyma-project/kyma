package metrics

import "time"

type ClusterInfo struct {
	Resources []NodeResources     `json:"resources"`
	Usage     []NodeMetrics       `json:"usage"`
	Volumes   []PersistentVolumes `json:"persistentVolumes"`
}

type NodeResources struct {
	Name         string       `json:"nodeName"`
	InstanceType string       `json:"instanceType"`
	Capacity     NodeCapacity `json:"capacity"`
}

type NodeCapacity struct {
	CPU              string `json:"cpu"`
	EphemeralStorage string `json:"ephemeralStorage"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
}

type NodeMetrics struct {
	Name                     string    `json:"nodeName"`
	StartCollectingTimestamp time.Time `json:"startCollectingTimestamp"`
	Usage                    NodeUsage `json:"usage"`
}

type NodeUsage struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`

	// Deprecated: It's always set to 0.
	EphemeralStorage string `json:"ephemeralStorage"`
	// Deprecated: It's always set to 0.
	Pods string `json:"pods"`
}

type PersistentVolumes struct {
	Name     string `json:"name"`
	Capacity string `json:"capacity"`
	Claim    *Claim `json:"claim,omitempty"`
}

type Claim struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
