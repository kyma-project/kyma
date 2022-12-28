package compass_runtime_agent

import "fmt"

type config struct {
	DirectorURL                       string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipDirectorCertVerification      bool   `envconfig:"default=false"`
	OAuthCredentialsNamespace         string `envconfig:"default=test"`
	SystemNamespace                   string `envconfig:"default=kyma-system"`
	CompassSystemNamespace            string `envconfig:"default=compass-system"`
	CompassRuntimeAgentDeploymentName string `envconfig:"default=compass-runtime-agent"`
	OAuthCredentialsSecretName        string `envconfig:"default=oauth-compass-credentials"`
	TestingTenant                     string `envconfig:"default=tenant"`
}

func (c *config) String() string {
	return fmt.Sprintf("DirectorURL: %s, SkipDirectorCertVerification: %v, OAuthCredentialsNamespace: %s, IntegrationNamespace: %s, CompassSystemNamespace: %s, OAuthCredentialsSecretName: %s, TestingTenant %s",
		c.DirectorURL, c.SkipDirectorCertVerification, c.OAuthCredentialsNamespace, c.SystemNamespace, c.CompassSystemNamespace, c.OAuthCredentialsSecretName, c.TestingTenant)
}
