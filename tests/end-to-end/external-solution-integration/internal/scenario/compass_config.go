package scenario

type CompassEnvConfig struct {
	Tenant             string
	ScenariosLabelKey  string `envconfig:"default=scenarios"`
	DexSecretName      string
	DexSecretNamespace string
	RuntimeID          string
}

// GetRuntimeID returns Compass ID of runtime that is tested
func (s *CompassEnvConfig) GetRuntimeID() string {
	return s.RuntimeID
}

// GetDexSecret returns name and namespace of secret with dex account
func (s *CompassEnvConfig) GetDexSecret() (string, string) {
	return s.DexSecretName, s.DexSecretNamespace
}

// GetScenariosLabelKey returns Compass label key for scenarios label
func (s *CompassEnvConfig) GetScenariosLabelKey() string {
	return s.ScenariosLabelKey
}

// GetDefaultTenant returns Compass ID of tenant that is used for tests
func (s *CompassEnvConfig) GetDefaultTenant() string {
	return s.Tenant
}
