// +build acceptance

package servicecatalogaddons

import (
	"fmt"
	"math/rand"
	"testing"

	"time"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type addonsConfigurationEvent struct {
	Type                string
	AddonsConfiguration shared.AddonsConfiguration
}

type addonsConfigurations struct {
	AddonsConfigurations []shared.AddonsConfiguration
}

type createAddonsConfigurationResponse struct {
	CreateAddonsConfiguration shared.AddonsConfiguration
}

type deleteAddonsConfigurationResponse struct {
	DeleteAddonsConfiguration shared.AddonsConfiguration
}

type addAddonsConfigurationURLsResponse struct {
	AddAddonsConfigurationURLs shared.AddonsConfiguration
}

type removeAddonsConfigurationURLsResponse struct {
	RemoveAddonsConfigurationURLs shared.AddonsConfiguration
}

func TestAddonsConfigurationMutationsAndQueries(t *testing.T) {
	// GIVEN
	suite := newAddonsConfigurationSuite(t)

	// WHEN
	t.Log("Subscribe Addons Configuration")
	subscription := suite.subscribeAddonsConfiguration()
	defer subscription.Close()

	t.Log("Create Addons Configuration")
	var res createAddonsConfigurationResponse
	err := suite.gqlCli.Do(suite.fixCreateAddonsConfigurationsRequest(), &res)

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, res.CreateAddonsConfiguration)

	// WHEN
	event, err := suite.readAddonsConfigurationEvent(subscription)

	// THEN
	t.Log("Check subscription event")
	assert.NoError(t, err)
	assert.True(t, event.Type == "UPDATE" || event.Type == "ADD")
	assert.Equal(t, suite.givenAddonsConfiguration.Name, event.AddonsConfiguration.Name)

	// WHEN
	t.Log("Query Addons Configuration")
	var query addonsConfigurations
	err = suite.gqlCli.Do(suite.fixAddonsConfigurationRequest(), &query)

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfigurationList(suite.givenAddonsConfiguration, query.AddonsConfigurations)

	expURL := "newURL"
	var addRes addAddonsConfigurationURLsResponse
	err = suite.gqlCli.Do(suite.fixAddRepoAddonsConfigurationRequest([]string{expURL}), &addRes)
	assert.NoError(t, err)
	assert.Contains(t, addRes.AddAddonsConfigurationURLs.Urls, expURL)

	var rmRes removeAddonsConfigurationURLsResponse
	err = suite.gqlCli.Do(suite.fixRemoveRepoAddonsConfigurationRequest([]string{expURL}), &rmRes)
	assert.NoError(t, err)
	assert.NotContains(t, rmRes.RemoveAddonsConfigurationURLs.Urls, expURL)

	t.Log("Delete Addons Configuration")
	var deleteRes deleteAddonsConfigurationResponse
	err = suite.gqlCli.Do(suite.fixDeleteAddonsConfigurationRequest(), &deleteRes)
	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, deleteRes.DeleteAddonsConfiguration)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {suite.fixAddonsConfigurationRequest()},
		auth.Create: {suite.fixCreateAddonsConfigurationsRequest()},
		auth.Delete: {suite.fixDeleteAddonsConfigurationRequest()},
		auth.Watch:  {suite.fixAddonConfigurationSubscription()},
	}
	AuthSuite.Run(t, ops)
}

func generateRandomName() string {
	var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "abcdefghijklmnopqrstuvwxyz"
	const length = 8

	str := make([]byte, length)
	for i := range str {
		str[i] = charset[random.Intn(len(charset))]
	}
	return string(str)
}

func newAddonsConfigurationSuite(t *testing.T) *addonsConfigurationTestSuite {
	c, err := graphql.New()
	require.NoError(t, err)
	addonsCli, _, err := client.NewAddonsConfigurationsClientWithConfig()
	require.NoError(t, err)

	name := generateRandomName()
	return &addonsConfigurationTestSuite{
		gqlCli:                   c,
		addonsCli:                addonsCli,
		t:                        t,
		givenAddonsConfiguration: fixture.AddonsConfiguration(name, []string{"test"}, map[string]string{"label": "true"}),
	}
}

type addonsConfigurationTestSuite struct {
	gqlCli    *graphql.Client
	addonsCli *versioned.Clientset
	t         *testing.T

	givenAddonsConfiguration shared.AddonsConfiguration
}

func (s *addonsConfigurationTestSuite) assertEqualAddonsConfiguration(expected shared.AddonsConfiguration, actual shared.AddonsConfiguration) {
	assert.Equal(s.t, expected.Name, actual.Name)
	assert.Equal(s.t, expected.Urls, actual.Urls)
	assert.Equal(s.t, expected.Labels, actual.Labels)
}

func (s *addonsConfigurationTestSuite) assertEqualAddonsConfigurationList(expected shared.AddonsConfiguration, actual []shared.AddonsConfiguration) {
	exist := false
	for _, addons := range actual {
		if addons.Name == expected.Name {
			s.assertEqualAddonsConfiguration(expected, addons)
			exist = true
		}
	}
	assert.True(s.t, exist)
}

func (s *addonsConfigurationTestSuite) fixCreateAddonsConfigurationsRequest() *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $namespace: String!, $urls: [String!]!, $labels: Labels!) {
			createAddonsConfiguration(name: $name, namespace: $namespace, urls: $urls, labels: $labels){
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)
	req.SetVar("urls", s.givenAddonsConfiguration.Urls)
	req.SetVar("labels", s.givenAddonsConfiguration.Labels)

	return req
}

func (s *addonsConfigurationTestSuite) fixDeleteAddonsConfigurationRequest() *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $namespace: String!) {
			deleteAddonsConfiguration(name: $name, namespace: $namespace) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)

	return req
}

func (s *addonsConfigurationTestSuite) fixAddRepoAddonsConfigurationRequest(newRepos []string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $namespace: String!, $urls: [String!]!) {
			addAddonsConfigurationURLs(name: $name, namespace: $namespace, urls: $urls) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)
	req.SetVar("urls", newRepos)

	return req
}

func (s *addonsConfigurationTestSuite) fixRemoveRepoAddonsConfigurationRequest(removeRepos []string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $namespace: String!, $urls: [String!]!) {
			removeAddonsConfigurationURLs(name: $name, namespace: $namespace, urls: $urls) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)
	req.SetVar("urls", removeRepos)

	return req
}

func (s *addonsConfigurationTestSuite) fixAddonsConfigurationRequest() *graphql.Request {
	query := fmt.Sprintf(`
		query ($namespace: String!) {
			addonsConfigurations(namespace: $namespace) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("namespace", TestNamespace)

	return req
}

func (s *addonsConfigurationTestSuite) addonsConfigurationDetailsFields() string {
	return `
		name
		urls
		labels
		status {
			phase
			repositories {
				url
				status
				addons {
					name
					version
					status
					message
					reason
				}
			}
		}
	`
}

func (s *addonsConfigurationTestSuite) fixAddonConfigurationSubscription() *graphql.Request {
	query := fmt.Sprintf(`
			subscription ($namespace: String!) {
				addonsConfigurationEvent(namespace: $namespace) {
					%s
				}
			}
		`, s.addonsConfigurationEventDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("namespace", TestNamespace)
	return req
}

func (s *addonsConfigurationTestSuite) subscribeAddonsConfiguration() *graphql.Subscription {
	return s.gqlCli.Subscribe(s.fixAddonConfigurationSubscription())
}

func (s *addonsConfigurationTestSuite) readAddonsConfigurationEvent(sub *graphql.Subscription) (addonsConfigurationEvent, error) {
	type Response struct {
		AddonsConfigurationEvent addonsConfigurationEvent
	}
	var addonsEvent Response
	err := sub.Next(&addonsEvent, time.Second*10)

	return addonsEvent.AddonsConfigurationEvent, err
}

func (s *addonsConfigurationTestSuite) addonsConfigurationEventDetailsFields() string {
	return `
        type
        addonsConfiguration {
			name
        }
    `
}
