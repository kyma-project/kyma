// +build acceptance

package application

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type applicationMutationOutput struct {
	Name        string
	Description string
	Labels      map[string]string
}

type application struct {
	Name                string
	Description         string
	Labels              map[string]string
	Services            []applicationService
	EnabledInNamespaces []string
	Status              string
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
	c, err := graphql.New()
	require.NoError(t, err)

	module.SkipPluggableTestIfShould(t, c, ModuleName)

	t.Log("Subscribe On Applications")
	subscription := subscribeApplicationEvent(c)
	defer subscription.Close()

	const fixName = "test-console-backend-application"
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

	t.Log("Update Application")
	fixApp = fixApplication(fixName, "desc2", map[string]string{"lab": "fix"})

	updateResp, err := updateApplicationMutation(c, fixApp)
	require.NoError(t, err)
	checkApplicationMutationOutput(t, fixApp, updateResp.AppUpdateMutation)

	t.Log("Delete Application")
	deleteResp, err := deleteApplicationMutation(c, fixName)
	require.NoError(t, err)
	assert.Equal(t, fixName, deleteResp.ApplicationDeleteMutation.Name)

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.Create: {fixCreateApplicationMutation(fixApplication("", "", map[string]string{}))},
		auth.Update: {fixUpdateApplicationMutation(fixApp)},
		auth.Delete: {fixDeleteApplicationMutation(fixName)},
	}
	as.Run(t, ops)
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
		enabledInNamespaces
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
	req := fixCreateApplicationMutation(given)
	var response appCreateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func updateApplicationMutation(c *graphql.Client, app *application) (appUpdateMutationResponse, error) {
	req := fixUpdateApplicationMutation(app)
	var response appUpdateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func deleteApplicationMutation(c *graphql.Client, appName string) (appDeleteMutationResponse, error) {
	req := fixDeleteApplicationMutation(appName)

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

func fixCreateApplicationMutation(app *application) *graphql.Request {
	query := fmt.Sprintf(`
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				createApplication(name: $name, description: $description, labels: $labels) {
					%s
				}
			}
	`, appMutationOutputFields())

	req := graphql.NewRequest(query)
	req.SetVar("name", app.Name)
	req.SetVar("description", app.Description)
	req.SetVar("labels", app.Labels)

	return req
}

func fixUpdateApplicationMutation(app *application) *graphql.Request {
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

	return req
}

func fixDeleteApplicationMutation(appName string) *graphql.Request {
	query := `
			mutation ($name: String!) {
				deleteApplication(name: $name) {
					name
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", appName)

	return req
}
