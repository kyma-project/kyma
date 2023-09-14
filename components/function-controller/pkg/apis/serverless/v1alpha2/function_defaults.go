package v1alpha2

const DefaultingConfigKey = "defaulting-config"

type ResourcesPreset struct {
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

type FunctionResourcesDefaulting struct {
	DefaultPreset  string
	Presets        map[string]ResourcesPreset
	RuntimePresets map[string]string
}

type BuildJobResourcesDefaulting struct {
	DefaultPreset string
	Presets       map[string]ResourcesPreset
}

type FunctionDefaulting struct {
	Resources FunctionResourcesDefaulting
}

type BuildJobDefaulting struct {
	Resources BuildJobResourcesDefaulting
}

type DefaultingConfig struct {
	Function FunctionDefaulting
	BuildJob BuildJobDefaulting
	Runtime  Runtime
}
