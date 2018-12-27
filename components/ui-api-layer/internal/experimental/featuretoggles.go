package experimental

type FeatureToggles struct {
	ModulePluggability bool `envconfig:"default=false,MODULE_PLUGGABILITY"`
}
