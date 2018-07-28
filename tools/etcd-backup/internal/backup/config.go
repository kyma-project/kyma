package backup

// Config holds Executor configuration
type Config struct {
	// EtcdEndpoints specifies the endpoints of an etcd cluster.
	// When multiple endpoints are given, the backup operator retrieves
	// the backup from the endpoint that has the most up-to-date state.
	// The given endpoints must belong to the same etcd cluster.
	EtcdEndpoints []string

	// ConfigMapNameForTracing is the name of the k8s ConfigMap where the
	// path to the ABS backup is saved (only from the last success).
	ConfigMapNameForTracing string
}
