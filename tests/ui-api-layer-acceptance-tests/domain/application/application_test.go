// +build acceptance

package application

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type applicationMutationOutput struct {
	Name        string
	Description string
	Labels      map[string]string
}

type application struct {
	Name                  string
	Description           string
	Labels                map[string]string
	Services              []applicationService
	EnabledInEnvironments []string
	Status                string
}

type applicationService struct {
	Id                  string
	DisplayName         string
	LongDescription     string
	ProviderDisplayName string
	Tags                string
	Entries             applicationEntry
}

type applicationEntry struct {
	Type        string
	GatewayUrl  string
	AccessLabel string
}

type appCreateMutationResponse struct {
	AppCreateMutation applicationMutationOutput `json:"createApplication"`
}

type appUpdateMutationResponse struct {
	AppUpdateMutation applicationMutationOutput `json:"updateApplication"`
}

type appDeleteMutationResponse struct {
	ApplicationDeleteMutation applicationMutationOutput `json:"deleteApplication"`
}

type applicationEvent struct {
	Type        string
	Application application
}

func TestApplicationMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}
	c, err := graphql.New()
	require.NoError(t, err)

	t.Log("Subscribe On Applications")
	subscription := subscribeApplicationEvent(c)
	defer subscription.Close()

	const fixName = "test-ui-api-application"
	fixApp := fixApplication(fixName, "fix-desc1", map[string]string{"fix": "lab"})

	t.Log("Create Application")
	resp, err := createApplicationMutation(c, fixApp)
	require.NoError(t, err)
	checkApplicationMutationOutput(t, fixApp, resp.AppCreateMutation)

	t.Log("Check Subscription Event")
	expectedEvent := fixApplicationEvent("ADD", fixApp)
	event, err := readApplicationEvent(subscription)
	assert.NoError(t, err)
	checkApplicationEvent(t, expectedEvent, event)

	defer func() {
		t.Log("Delete Application")
		deleteResp, err := deleteApplicationMutation(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, fixName, deleteResp.ApplicationDeleteMutation.Name)
	}()

	t.Log("Update Application")
	fixApp = fixApplication(fixName, "desc2", map[string]string{"lab": "fix"})

	updateResp, err := updateApplicationMutation(c, fixApp)
	require.NoError(t, err)
	checkApplicationMutationOutput(t, fixApp, updateResp.AppUpdateMutation)
}

func readApplicationEvent(sub *graphql.Subscription) (applicationEvent, error) {
	type Response struct {
		ApplicationEvent applicationEvent
	}
	var appEvent Response
	err := sub.Next(&appEvent, tester.DefaultSubscriptionTimeout)

	return appEvent.ApplicationEvent, err
}

func subscribeApplicationEvent(c *graphql.Client) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription {
				%s
			}
		`, appEventFields())
	req := graphql.NewRequest(query)
	return c.Subscribe(req)
}

func appEventFields() string {
	return fmt.Sprintf(`
        applicationEvent {
			type
			application{
				%s
			}
        }
    `, appFields())
}

func appMutationOutputFields() string {
	return `
		name
		description
		labels
    `
}

func appFields() string {
	return `
		name
		description
		labels
		enabledInEnvironments
		status               
		services {
			id                 
			displayName        
			longDescription    
			providerDisplayName
			tags               
			entries {
				type       
				gatewayUrl 
				accessLabel
			}         
		}          
    `
}

func checkApplicationEvent(t *testing.T, expected, actual applicationEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Application.Name, actual.Application.Name)
}

func checkApplicationMutationOutput(t *testing.T, app *application, appMutation applicationMutationOutput) {
	assert.Equal(t, app.Name, appMutation.Name)
	assert.Equal(t, app.Description, appMutation.Description)
	assert.Equal(t, app.Labels, appMutation.Labels)
}

func createApplicationMutation(c *graphql.Client, given *application) (appCreateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				createApplication(name: $name, description: $description, labels: $labels) {
					%s
				}
			}
	`, appMutationOutputFields())

	req := graphql.NewRequest(query)
	req.SetVar("name", given.Name)
	req.SetVar("description", given.Description)
	req.SetVar("labels", given.Labels)
	var response appCreateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func updateApplicationMutation(c *graphql.Client, app *application) (appUpdateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				updateApplication(name: $name, description: $description, labels: $labels) {
					%s
				}
			}
	`, appMutationOutputFields())

	req := graphql.NewRequest(query)
	req.SetVar("name", app.Name)
	req.SetVar("description", app.Description)
	req.SetVar("labels", app.Labels)
	var response appUpdateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func deleteApplicationMutation(c *graphql.Client, appName string) (appDeleteMutationResponse, error) {
	query := `
			mutation ($name: String!) {
				deleteApplication(name: $name) {
					name
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", appName)
	var response appDeleteMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func fixApplication(name string, desc string, labels map[string]string) *application {
	return &application{
		Name:        name,
		Description: desc,
		Labels:      labels,
	}
}

func fixApplicationEvent(eventType string, app *application) applicationEvent {
	return applicationEvent{
		Type:        eventType,
		Application: *app,
	}
}
