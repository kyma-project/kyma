package experimental

// FeatureToggles stores toggles for experimental features
type FeatureToggles struct {
	// AddonsConfigurationFeatureEnabled when enabled then it replace the ConfigMap implementation with the ClusterAddonsConfiguration CR.
	AddonsConfigurationFeatureEnabled bool `envconfig:"default=false"`
}
