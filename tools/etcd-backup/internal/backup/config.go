package backup

// Config holds Executor configuration
type Config struct {
	// EtcdEndpoints specifies the endpoints of an etcd cluster.
	// When multiple endpoints are given, the backup operator retrieves
	// the backup from the endpoint that has the most up-to-date state.
	// The given endpoints must belong to the same etcd cluster.
	EtcdEndpoints []string

	// ConfigMapNameForTracing is the name of the ConfigMap where
	// the path to the last successful ABS backup is saved.
	ConfigMapNameForTracing string
}
