package experimental

// FeatureToggles stores toggles for experimental features
type FeatureToggles struct {
	// ClusterAddonsConfigurationCRDEnabled when enabled then it replace the ConfigMap implementation with the ClusterAddonsConfiguration CR.
	ClusterAddonsConfigurationCRDEnabled bool `envconfig:"default=false"`
}
