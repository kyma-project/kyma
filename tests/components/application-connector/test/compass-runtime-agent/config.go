package compass_runtime_agent

import "fmt"

type config struct {
	DirectorURL                       string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipDirectorCertVerification      bool   `envconfig:"default=false"`
	TestNamespace                     string `envconfig:"default=test"`
	IntegrationNamespace              string `envconfig:"default=kyma-integration"`
	CompassSystemNamespace            string `envconfig:"default=compass-system"`
	CompassRuntimeAgentDeploymentName string `envconfig:"default=compass-runtime-agent"`
	OauthCredentialsSecretName        string `envconfig:"default=oauth-compass-credentials"`
	TestingTenant                     string `envconfig:"default=tenant"`
	KubeconfigPath                    string
}

func (c *config) String() string {
	return fmt.Sprintf("DirectorURL: %s, SkipDirectorCertVerification: %v, TestCredentialsNamespace: %s, OauthCredentialsSecretName: %s, TestingTenant %s",
		c.DirectorURL, c.SkipDirectorCertVerification, c.TestNamespace, c.OauthCredentialsSecretName, c.TestingTenant)
}
