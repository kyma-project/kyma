package etcd

import "github.com/coreos/etcd/clientv3"

// TODO list:
// - Use etcd lease for garbage collection of removed elements.
//   Create lease on element delete and attach it to each object which should be deleted.

const (
	entityNamespaceSeparator = "/"

	entityNamespaceBundle          = "bundle"
	entityNamespaceBundleMappingID = "id"
	entityNamespaceBundleMappingNV = "nv"

	entityNamespaceChart             = "chart"
	entityNamespaceInstance          = "instance"
	entityNamespaceInstanceOperation = "instanceOperation"
	entityNamespaceInstanceBindData  = "instanceBindData"
)

// Config holds configuration for etcd access in storage.
type Config struct {
	Endpoints []string `json:"endpoints"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`

	ForceClient *clientv3.Client
}

func entityNamespacePrefixParts() []string {
	return []string{"helm-broker", "entity"}
}

// generic is a foundation for all drivers using etcd as storage.
type generic struct {
	kv clientv3.KV
}
