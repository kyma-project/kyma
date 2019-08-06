// +build acceptance

package servicecatalogaddons

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
)

type clusterAddonsConfigurationEvent struct {
	Type                string
	AddonsConfiguration shared.AddonsConfiguration
}

type clusterAddonsConfigurations struct {
	ClusterAddonsConfigurations []shared.AddonsConfiguration
}

type createClusterAddonsConfigurationResponse struct {
	CreateClusterAddonsConfiguration shared.AddonsConfiguration
}

type deleteClusterAddonsConfigurationResponse struct {
	DeleteClusterAddonsConfiguration shared.AddonsConfiguration
}

type addClusterAddonsConfigurationURLsResponse struct {
	AddClusterAddonsConfigurationURLs shared.AddonsConfiguration
}

type removeClusterAddonsConfigurationURLsResponse struct {
	RemoveClusterAddonsConfigurationURLs shared.AddonsConfiguration
}

func TestClusterAddonsConfigurationMutationsAndQueries(t *testing.T) {
	// GIVEN
	suite := newAddonsConfigurationSuite(t)

	t.Log("Subscribe Cluster Addons Configuration")
	subscription := suite.subscribeClusterAddonsConfiguration()
	defer subscription.Close()

	// WHEN
	t.Log("Create Cluster Addons Configuration")
	var createRes createClusterAddonsConfigurationResponse
	err := suite.gqlCli.Do(suite.fixCreateClusterAddonsConfigurationsRequest(), &createRes)

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, createRes.CreateClusterAddonsConfiguration)

	// WHEN
	event, err := suite.readClusterAddonsConfigurationEvent(subscription)

	// THEN
	t.Log("Check subscription event")
	assert.NoError(t, err)
	assert.True(t, event.Type == "UPDATE" || event.Type == "ADD")
	assert.Equal(t, suite.givenAddonsConfiguration.Name, event.AddonsConfiguration.Name)

	// WHEN
	t.Log("Query Cluster Addons Configuration")
	var res clusterAddonsConfigurations
	err = suite.gqlCli.Do(suite.fixClusterAddonsConfigurationRequest(), &res)

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfigurationList(suite.givenAddonsConfiguration, res.ClusterAddonsConfigurations)

	expURL := "newURL"
	var addRes addClusterAddonsConfigurationURLsResponse
	err = suite.gqlCli.Do(suite.fixAddRepoClusterAddonsConfigurationRequest([]string{expURL}), &addRes)
	assert.NoError(t, err)
	assert.Contains(t, addRes.AddClusterAddonsConfigurationURLs.Urls, expURL)

	var rmRes removeClusterAddonsConfigurationURLsResponse
	err = suite.gqlCli.Do(suite.fixRemoveRepoClusterAddonsConfigurationRequest([]string{expURL}), &addRes)
	assert.NoError(t, err)
	assert.NotContains(t, rmRes.RemoveClusterAddonsConfigurationURLs.Urls, expURL)

	t.Log("Delete Cluster Addons Configuration")

	var deleteRes deleteClusterAddonsConfigurationResponse
	err = suite.gqlCli.Do(suite.fixDeleteClusterAddonsConfigurationRequest(), &deleteRes)

	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, deleteRes.DeleteClusterAddonsConfiguration)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {suite.fixClusterAddonsConfigurationRequest()},
		auth.Create: {suite.fixCreateClusterAddonsConfigurationsRequest()},
		auth.Delete: {suite.fixDeleteClusterAddonsConfigurationRequest()},
	}
	AuthSuite.Run(t, ops)
}

func (s *addonsConfigurationTestSuite) fixCreateClusterAddonsConfigurationsRequest() *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $urls: [String!]!, $labels: Labels!) {
			createClusterAddonsConfiguration(name: $name, urls: $urls, labels: $labels){
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("urls", s.givenAddonsConfiguration.Urls)
	req.SetVar("labels", s.givenAddonsConfiguration.Labels)

	return req
}

func (s *addonsConfigurationTestSuite) fixDeleteClusterAddonsConfigurationRequest() *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!) {
			deleteClusterAddonsConfiguration(name: $name) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)

	return req
}

func (s *addonsConfigurationTestSuite) fixAddRepoClusterAddonsConfigurationRequest(newRepos []string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $urls: [String!]!) {
			addClusterAddonsConfigurationURLs(name: $name, urls: $urls) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("urls", newRepos)

	return req
}

func (s *addonsConfigurationTestSuite) fixRemoveRepoClusterAddonsConfigurationRequest(removeRepos []string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $urls: [String!]!) {
			removeClusterAddonsConfigurationURLs(name: $name, urls: $urls) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("urls", removeRepos)

	return req
}

func (s *addonsConfigurationTestSuite) fixClusterAddonsConfigurationRequest() *graphql.Request {
	query := fmt.Sprintf(`
		query {
			clusterAddonsConfigurations {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)

	return req
}

func (s *addonsConfigurationTestSuite) subscribeClusterAddonsConfiguration() *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription {
				clusterAddonsConfigurationEvent {
					%s
				}
			}
		`, s.addonsConfigurationEventDetailsFields())
	req := graphql.NewRequest(query)

	return s.gqlCli.Subscribe(req)
}

func (s *addonsConfigurationTestSuite) readClusterAddonsConfigurationEvent(sub *graphql.Subscription) (clusterAddonsConfigurationEvent, error) {
	type Response struct {
		ClusterAddonsConfigurationEvent clusterAddonsConfigurationEvent
	}
	var addonsEvent Response
	err := sub.Next(&addonsEvent, time.Second*10)

	return addonsEvent.ClusterAddonsConfigurationEvent, err
}
