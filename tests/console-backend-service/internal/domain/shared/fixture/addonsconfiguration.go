package fixture

import "github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"

func AddonsConfiguration(name string, urls []string, labels map[string]string) shared.AddonsConfiguration {
	return shared.AddonsConfiguration{
		Name:   name,
		Urls:   urls,
		Labels: labels,
	}
}
