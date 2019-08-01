package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"testing"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/stretchr/testify/require"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AddonsConfigurationEvent struct {
	Type                string
	AddonsConfiguration shared.AddonsConfiguration
}

type addonsConfigurationMutationResponse struct {
	AddonsConfiguration shared.AddonsConfiguration
}

type addonsConfigurationQueryResponse struct {
	AddonsConfigurations []shared.AddonsConfiguration
}

func TestAddonsConfigurationMutationsAndQueries(t *testing.T) {
	// GIVEN
	suite := newAddonsConfigurationSuite(t)
	suite.createAddonsConfiguration()
	defer suite.deleteAddonsConfiguration()

	t.Log("Subscribe Addons Configuration")
	subscription := suite.subscribeAddonsConfiguration()
	defer subscription.Close()

	// WHEN
	t.Log("Create Addons Configuration")
	createRes, err := suite.addonsConfigurationRequest(suite.fixCreateAddonsConfigurationsRequest)

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, createRes.AddonsConfiguration)

	// WHEN
	event, err := suite.readAddonsConfigurationEvent(subscription)

	// THEN
	t.Log("Check subscription event")
	assert.NoError(t, err)
	suite.assertEqualAddonsConfigurationEvent(event)

	// WHEN
	t.Log("Query Addons Configuration")
	res, err := suite.queryAddonsConfiguration()

	// THEN
	assert.NoError(t, err)
	suite.assertEqualAddonsConfiguration(suite.givenAddonsConfiguration, res.AddonsConfigurations[0])

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {suite.fixAddonsConfigurationRequest()},
		auth.Create: {suite.fixCreateAddonsConfigurationsRequest()},
		auth.Delete: {suite.fixDeleteAddonsConfigurationRequest()},
	}
	auth.TestSuite{}.Run(t, ops)
}

func newAddonsConfigurationSuite(t *testing.T) *addonsConfigurationTestSuite {
	c, err := graphql.New()
	require.NoError(t, err)
	addonsCli, _, err := client.NewAddonsConfigurationsClientWithConfig()
	require.NoError(t, err)

	return &addonsConfigurationTestSuite{
		gqlCli:    c,
		addonsCli: addonsCli,
		t:         t,
	}
}

type addonsConfigurationTestSuite struct {
	gqlCli    *graphql.Client
	addonsCli *versioned.Clientset
	t         *testing.T

	givenAddonsConfiguration shared.AddonsConfiguration
}

func (s *addonsConfigurationTestSuite) createAddonsConfiguration() error {
	repos := make([]v1alpha1.SpecRepository, 0)
	for _, url := range s.givenAddonsConfiguration.Urls {
		repos = append(repos, v1alpha1.SpecRepository{URL: url})
	}

	_, err := s.addonsCli.AddonsV1alpha1().AddonsConfigurations(TestNamespace).Create(&v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.givenAddonsConfiguration.Name,
			Namespace: TestNamespace,
			Labels:    s.givenAddonsConfiguration.Labels,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{},
			},
		},
	})
	return err
}

func (s *addonsConfigurationTestSuite) fixCreateAddonsConfigurationsRequest() *graphql.Request {
	query := fmt.Sprintf(`
		mutation ($name: String!, $namespace: String!, urls: [String!]!, labels: Labels!) {
			createAddonsConfiguration(namespace: $namespace, name: $name, urls: $urls, labels: $labels){
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

func (s *addonsConfigurationTestSuite) addonsConfigurationRequest(r func() *graphql.Request) (addonsConfigurationMutationResponse, error) {
	var res addonsConfigurationMutationResponse
	err := s.gqlCli.Do(r(), &res)

	return res, err
}

func (s *addonsConfigurationTestSuite) fixDeleteAddonsConfigurationRequest() *graphql.Request {
	query := `
		mutation ($name: String!, $namespace: String!) {
			deleteAddonsConfiguration(name: $name, namespace: $namespace) {
				name
				namespace
			}
		}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)

	return req
}

func (s *addonsConfigurationTestSuite) deleteAddonsConfiguration() (addonsConfigurationQueryResponse, error) {
	req := s.fixDeleteAddonsConfigurationRequest()

	var res addonsConfigurationQueryResponse
	err := s.gqlCli.Do(req, &res)

	return res, err
}

func (s *addonsConfigurationTestSuite) assertEqualAddonsConfiguration(expected, actual shared.AddonsConfiguration) {
	assert.Equal(s.t, expected.Name, actual.Name)
	assert.NotEmpty(s.t, actual.Status)
}

func (s *addonsConfigurationTestSuite) fixAddonsConfigurationRequest() *graphql.Request {
	query := fmt.Sprintf(`
		query ($name: String!, $namespace: String!) {
			addonsConfiguration(name: $name, namespace: $namespace) {
				%s
			}
		}
	`, s.addonsConfigurationDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenAddonsConfiguration.Name)
	req.SetVar("namespace", TestNamespace)

	return req
}

func (s *addonsConfigurationTestSuite) queryAddonsConfiguration() (addonsConfigurationQueryResponse, error) {
	req := s.fixAddonsConfigurationRequest()

	var res addonsConfigurationQueryResponse
	err := s.gqlCli.Do(req, &res)

	return res, err
}

func (s *addonsConfigurationTestSuite) addonsConfigurationDetailsFields() string {
	return `
		name
	`
}

func (s *addonsConfigurationTestSuite) subscribeAddonsConfiguration() *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($namespace: String!) {
				addonsConfigurationEvent(namespace: $namespace) {
					%s
				}
			}
		`, s.addonsConfigurationEventDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("namespace", TestNamespace)

	return s.gqlCli.Subscribe(req)
}

func (s *addonsConfigurationTestSuite) readAddonsConfigurationEvent(sub *graphql.Subscription) (AddonsConfigurationEvent, error) {
	type Response struct {
		AddonsConfigurationEvent AddonsConfigurationEvent
	}
	var addonsEvent Response
	err := sub.Next(&addonsEvent, tester.DefaultSubscriptionTimeout)

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

func (s *addonsConfigurationTestSuite) assertEqualAddonsConfigurationEvent(event AddonsConfigurationEvent) {
	assert.Equal(s.t, "ADD", event.Type)
	assert.Equal(s.t, s.givenAddonsConfiguration.Name, event.AddonsConfiguration.Name)
}
