package helpers

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	gcli "github.com/machinebox/graphql"
)

type CompassFixtures struct {
	gqlFieldsProvider gql.GqlFieldsProvider
}

func NewCompassFixtures() CompassFixtures {
	return CompassFixtures{
		gqlFieldsProvider: gql.GqlFieldsProvider{},
	}
}

func (cf *CompassFixtures) FixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, cf.gqlFieldsProvider.ForApplication()))
}

func (cf *CompassFixtures) FixUnregisterApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterApplication(id: "%s") {
					%s
				}
			}`,
			applicationID, cf.gqlFieldsProvider.ForApplication()))
}

func (cf *CompassFixtures) FixRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, cf.gqlFieldsProvider.ForRuntime()))
}

func (cf *CompassFixtures) FixSetRuntimeLabelRequest(runtimeID, labelKey string, labelValue []string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
						%s
					}
				}`,
			runtimeID, labelKey, labelValue, cf.gqlFieldsProvider.ForLabel()))
}

func (cf *CompassFixtures) FixDeleteRuntimeLabelRequest(runtimeID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
					%s
				}
			}`, runtimeID, labelKey, cf.gqlFieldsProvider.ForLabel()))
}

func (cf *CompassFixtures) FixRequestOneTimeTokenForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
	result: requestOneTimeTokenForApplication(id: "%s") {
		%s
	}
}`, appID, cf.gqlFieldsProvider.ForOneTimeTokenForApplication()))
}

func (cf *CompassFixtures) FixGetApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, appID, cf.gqlFieldsProvider.ForApplication()))
}
