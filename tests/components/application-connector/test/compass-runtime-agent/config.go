package compass_runtime_agent

import "fmt"

type config struct {
	DirectorURL                  string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipDirectorCertVerification bool   `envconfig:"default=false"`
	TestCredentialsNamespace     string `envconfig:"default=test"`
	OauthCredentialsSecretName   string `envconfig:"default=oauth-compass-credentials"`
	TestingTenant                string `envconfig:"default=tenant"`
}

func (c *config) String() string {
	return fmt.Sprintf("DirectorURL: %s, SkipDirectorCertVerification: %v, TestCredentialsNamespace: %s, OauthCredentialsSecretName: %s, TestingTenant %s",
		c.DirectorURL, c.SkipDirectorCertVerification, c.TestCredentialsNamespace, c.OauthCredentialsSecretName, c.TestingTenant)
}
